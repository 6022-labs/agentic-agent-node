package middlewares

import (
	"strings"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	core_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/metrics/interfaces"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

const (
	PaymentRequiredHeader  = "PAYMENT-REQUIRED"
	PaymentSignatureHeader = "PAYMENT-SIGNATURE"
	PaymentResponseHeader  = "PAYMENT-RESPONSE"

	bearerPrefix = "Bearer "
)

// ResponseAccessGateMiddleware executes the runtime's verdict; it holds no policy of its own.
type ResponseAccessGateMiddleware struct {
	logger                     *zap.Logger
	jwtValidator               core_interfaces.JwtValidatorInterface
	paymentReceiveAccessClient gateway_interfaces.PaymentReceiveAccessClientInterface
	x402FacilitatorClient      gateway_interfaces.X402FacilitatorClientInterface
	paymentMetrics             metrics_interfaces.PaymentMetricsInterface
}

func NewResponseAccessGateMiddleware(
	logger *zap.Logger,
	jwtValidator core_interfaces.JwtValidatorInterface,
	paymentReceiveAccessClient gateway_interfaces.PaymentReceiveAccessClientInterface,
	x402FacilitatorClient gateway_interfaces.X402FacilitatorClientInterface,
	paymentMetrics metrics_interfaces.PaymentMetricsInterface,
) *ResponseAccessGateMiddleware {
	return &ResponseAccessGateMiddleware{
		logger:                     logger,
		jwtValidator:               jwtValidator,
		paymentReceiveAccessClient: paymentReceiveAccessClient,
		x402FacilitatorClient:      x402FacilitatorClient,
		paymentMetrics:             paymentMetrics,
	}
}

func (m *ResponseAccessGateMiddleware) Handle(ctx fiber.Ctx) error {
	decision, err := m.paymentReceiveAccessClient.Authorize(
		ctx.Context(),
		m.jwtAddress(ctx),
		ctx.Get(PaymentSignatureHeader),
	)
	if err != nil {
		// This route is unauthenticated and the runtime's text can carry schema detail.
		m.logger.Error("x402 access decision unavailable", zap.Error(err))
		return mvc.WriteError(
			ctx,
			apperrors.NewUnavailableError("access_decision_unavailable", "the access decision is unavailable"),
		)
	}

	switch decision.Outcome {
	case models.PaymentReceiveAccessOutcomeFree:
		ctx.SetContext(helpers.ContextWithCallerId(ctx.Context(), decision.Payer))
		return ctx.Next()
	case models.PaymentReceiveAccessOutcomePayable:
		return m.settle(ctx, decision)
	default:
		return m.paymentRequired(ctx, decision)
	}
}

// jwtAddress returns the bearer's proven address, or "" when there is no usable token.
func (m *ResponseAccessGateMiddleware) jwtAddress(ctx fiber.Ctx) string {
	header := ctx.Get(fiber.HeaderAuthorization)
	if !strings.HasPrefix(header, bearerPrefix) {
		return ""
	}
	token := strings.TrimSpace(header[len(bearerPrefix):])
	if len(token) == 0 {
		return ""
	}
	address, err := m.jwtValidator.Verify(token)
	if err != nil {
		return ""
	}
	return address.Hex()
}

func (m *ResponseAccessGateMiddleware) settle(
	ctx fiber.Ctx,
	decision *models.ExternalPaymentReceiveAccessDecision,
) error {
	settlement := decision.Settlement
	if settlement == nil {
		return mvc.WriteError(
			ctx,
			apperrors.NewUnavailableError("access_decision_unavailable", "payable decision carried no settlement"),
		)
	}
	payload, err := x402.DecodePaymentPayload(ctx.Get(PaymentSignatureHeader))
	if err != nil {
		m.logger.Error("payment signature the runtime accepted failed to decode", zap.Error(err))
		decision.ErrorCode = "malformed_payment_signature"
		return m.paymentRequired(ctx, decision)
	}

	// The payer takes the identity role the JWT fills on other routes.
	ctx.SetContext(helpers.ContextWithCallerId(ctx.Context(), decision.Payer))

	verifyResponse, err := m.x402FacilitatorClient.Verify(
		ctx.Context(),
		settlement.FacilitatorBaseUrl,
		payload,
		&settlement.Requirements,
	)
	if err != nil {
		m.logger.Error("x402 facilitator verify failed", zap.Error(err))
		return mvc.WriteError(
			ctx,
			apperrors.NewUnavailableError("facilitator_unavailable", "the payment facilitator is unavailable"),
		)
	}
	if !verifyResponse.IsValid {
		decision.ErrorCode = verifyResponse.InvalidReason
		return m.paymentRequired(ctx, decision)
	}

	if err := ctx.Next(); err != nil {
		return err
	}
	if ctx.Response().StatusCode() >= fiber.StatusBadRequest {
		return nil
	}

	settleResponse, err := m.x402FacilitatorClient.Settle(
		ctx.Context(),
		settlement.FacilitatorBaseUrl,
		payload,
		&settlement.Requirements,
	)
	if err != nil || !settleResponse.Success {
		m.paymentMetrics.RecordRejected(ctx.Context(), settlement.Network)
		if err != nil {
			m.logger.Error("x402 settlement failed", zap.Error(err))
		} else {
			m.logger.Error("x402 settlement rejected", zap.String("reason", settleResponse.ErrorReason))
		}
		// The caller wasn't charged, so they must not receive the response.
		ctx.Response().ResetBody()
		decision.ErrorCode = "settlement_failed"
		return m.paymentRequired(ctx, decision)
	}
	m.paymentMetrics.RecordSettled(ctx.Context(), settlement.Network)

	encoded, err := x402.EncodeBase64Json(settleResponse)
	if err != nil {
		m.logger.Error("failed to encode settle response", zap.Error(err))
		return nil
	}
	ctx.Set(PaymentResponseHeader, encoded)
	return nil
}

// paymentRequired wraps the runtime's advertised requirements in the spec 402.
func (m *ResponseAccessGateMiddleware) paymentRequired(
	ctx fiber.Ctx,
	decision *models.ExternalPaymentReceiveAccessDecision,
) error {
	accepts := decision.Accepts
	if accepts == nil {
		accepts = []x402.PaymentRequirements{}
	}
	body := &x402.PaymentRequired{
		X402Version: x402.Version,
		Resource: x402.ResourceInfo{
			Url:         ctx.BaseURL() + ctx.OriginalURL(),
			Description: "Agent inference turn",
			MimeType:    fiber.MIMEApplicationJSON,
		},
		Accepts: accepts,
		Error:   decision.ErrorCode,
	}
	if encoded, err := x402.EncodeBase64Json(body); err == nil {
		ctx.Set(PaymentRequiredHeader, encoded)
	}
	return ctx.Status(fiber.StatusPaymentRequired).JSON(body)
}
