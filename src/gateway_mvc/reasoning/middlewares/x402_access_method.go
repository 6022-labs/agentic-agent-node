package middlewares

import (
	"errors"
	"math/big"
	"regexp"
	"slices"
	"strings"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	x402_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/x402/services/interfaces"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway_mvc/metrics/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

const (
	PaymentRequiredHeader  = "PAYMENT-REQUIRED"
	PaymentSignatureHeader = "PAYMENT-SIGNATURE"
	PaymentResponseHeader  = "PAYMENT-RESPONSE"

	X402PayerLocalKey = "x402.payer"

	methodX402 = "x402"

	// x402Priority runs the payment link last (terminal fallback).
	x402Priority = 100

	// Rules carrying this pattern are public and advertised in 402s.
	everyoneAddressPattern = ".*"
)

// X402AccessMethod gates a request over x402 v2: the payer gets the cheapest
// rule whose patterns match (price "0" free, never settled), else refused.
type X402AccessMethod struct {
	logger                   *zap.Logger
	paymentFacilitatorClient gateway_interfaces.PaymentFacilitatorClientInterface
	x402PaymentVerifier      x402_interfaces.X402PaymentVerifierInterface
	x402FacilitatorClient    x402_interfaces.X402FacilitatorClientInterface
	paymentMetrics           metrics_interfaces.PaymentMetricsInterface
}

func NewX402AccessMethod(
	logger *zap.Logger,
	paymentFacilitatorClient gateway_interfaces.PaymentFacilitatorClientInterface,
	x402PaymentVerifier x402_interfaces.X402PaymentVerifierInterface,
	x402FacilitatorClient x402_interfaces.X402FacilitatorClientInterface,
	paymentMetrics metrics_interfaces.PaymentMetricsInterface,
) *X402AccessMethod {
	return &X402AccessMethod{
		logger:                   logger,
		paymentFacilitatorClient: paymentFacilitatorClient,
		x402PaymentVerifier:      x402PaymentVerifier,
		x402FacilitatorClient:    x402FacilitatorClient,
		paymentMetrics:           paymentMetrics,
	}
}

// x402Rule pairs a rule with its decoded x402 config.
type x402Rule struct {
	rule   *models.ExternalPaymentReceiveRule
	config *x402.X402Config
}

func (m *X402AccessMethod) Priority() int {
	return x402Priority
}

// x402 is terminal: it always handles, settling a payment or advertising a 402.
func (m *X402AccessMethod) TryHandle(ctx fiber.Ctx, rules []models.ExternalPaymentReceiveRule) (bool, error) {
	return true, m.handle(ctx, rules)
}

func (m *X402AccessMethod) MatchesFreeRule(rules []models.ExternalPaymentReceiveRule, address common.Address) bool {
	for _, candidate := range m.parseX402Rules(rules) {
		if m.matchesAddressPatterns(candidate.config.AddressPatterns, address) && m.isZeroPrice(candidate.rule.Price) {
			return true
		}
	}
	return false
}

func (m *X402AccessMethod) handle(ctx fiber.Ctx, rules []models.ExternalPaymentReceiveRule) error {
	// A rule is settleable only if its network has a facilitator.
	facilitators, err := m.paymentFacilitatorClient.Get(ctx.Context())
	if err != nil {
		return mvc.WriteError(ctx, apperrors.NewUnavailableError("payment_facilitators_unavailable", err.Error()))
	}
	x402Rules := m.parseX402Rules(rules)

	header := ctx.Get(PaymentSignatureHeader)
	if len(header) == 0 {
		return m.paymentRequired(ctx, x402Rules, facilitators, "")
	}

	payload, err := x402.DecodePaymentPayload(header)
	if err != nil {
		return m.paymentRequired(ctx, x402Rules, facilitators, "malformed_payment_signature")
	}

	candidates := m.matchRules(x402Rules, payload)
	if len(candidates) == 0 {
		return m.paymentRequired(ctx, x402Rules, facilitators, "no_matching_payment_requirements")
	}

	// Candidates share network+payTo but differ in asset (EIP-712 domain), so recover per candidate.
	payer, recovered, err := m.recoverAgainstCandidates(payload, candidates)
	if errors.Is(err, x402_interfaces.ErrX402AuthorizationExpired) {
		return m.paymentRequired(ctx, x402Rules, facilitators, "authorization_expired")
	}
	if err != nil {
		return m.paymentRequired(ctx, x402Rules, facilitators, "invalid_payment_signature")
	}

	// The payer takes the identity role the JWT fills on other routes.
	ctx.Locals(X402PayerLocalKey, payer.Hex())
	ctx.SetContext(helpers.ContextWithCallerId(ctx.Context(), payer.Hex()))

	entitled := m.cheapestEntitledRule(recovered, payer, facilitators)
	if entitled == nil {
		return m.paymentRequired(ctx, x402Rules, facilitators, "no_matching_payment_requirements")
	}

	if m.isZeroPrice(entitled.rule.Price) {
		return ctx.Next()
	}
	if !m.coversPrice(payload, entitled.rule) {
		return m.paymentRequired(ctx, x402Rules, facilitators, "insufficient_amount")
	}

	facilitator := facilitators[entitled.config.Network]
	requirements := m.toPaymentRequirements(entitled, facilitator.MaxTimeoutSeconds)
	verifyResponse, err := m.x402FacilitatorClient.Verify(ctx.Context(), facilitator.BaseUrl, payload, requirements)
	if err != nil {
		return mvc.WriteError(ctx, apperrors.NewUnavailableError("facilitator_unavailable", err.Error()))
	}
	if !verifyResponse.IsValid {
		return m.paymentRequired(ctx, x402Rules, facilitators, verifyResponse.InvalidReason)
	}

	if err := ctx.Next(); err != nil {
		return err
	}
	if ctx.Response().StatusCode() >= fiber.StatusBadRequest {
		return nil
	}

	settleResponse, err := m.x402FacilitatorClient.Settle(ctx.Context(), facilitator.BaseUrl, payload, requirements)
	if err != nil || !settleResponse.Success {
		m.paymentMetrics.RecordRejected(ctx.Context(), entitled.config.Network)
		if err != nil {
			m.logger.Error("x402 settlement failed", zap.Error(err))
		} else {
			m.logger.Error("x402 settlement rejected", zap.String("reason", settleResponse.ErrorReason))
		}
		// The caller wasn't charged, so they must not receive the response.
		ctx.Response().ResetBody()
		return m.paymentRequired(ctx, x402Rules, facilitators, "settlement_failed")
	}
	m.paymentMetrics.RecordSettled(ctx.Context(), entitled.config.Network)

	encoded, err := x402.EncodeBase64Json(settleResponse)
	if err != nil {
		m.logger.Error("failed to encode settle response", zap.Error(err))
		return nil
	}
	ctx.Set(PaymentResponseHeader, encoded)
	return nil
}

func (m *X402AccessMethod) recoverAgainstCandidates(
	payload *x402.PaymentPayload,
	candidates []x402Rule,
) (common.Address, []x402Rule, error) {
	var payer common.Address
	var recovered []x402Rule
	var lastErr error
	for _, candidate := range candidates {
		p, err := m.x402PaymentVerifier.RecoverPayer(payload, candidate.config)
		if errors.Is(err, x402_interfaces.ErrX402AuthorizationExpired) {
			return common.Address{}, nil, err
		}
		if err != nil {
			lastErr = err
			continue
		}
		payer = p
		recovered = append(recovered, candidate)
	}
	if len(recovered) == 0 {
		return common.Address{}, nil, lastErr
	}
	return payer, recovered, nil
}

// paymentRequired writes the spec 402, advertising only public rules (scoped ones would leak patterns).
func (m *X402AccessMethod) paymentRequired(
	ctx fiber.Ctx,
	x402Rules []x402Rule,
	facilitators map[string]models.ExternalPaymentFacilitator,
	errorCode string,
) error {
	accepts := []x402.PaymentRequirements{}
	for i := range x402Rules {
		if slices.Contains(x402Rules[i].config.AddressPatterns, everyoneAddressPattern) &&
			!m.unsettleablePaidRule(&x402Rules[i], facilitators) {
			maxTimeoutSeconds := facilitators[x402Rules[i].config.Network].MaxTimeoutSeconds
			accepts = append(accepts, *m.toPaymentRequirements(&x402Rules[i], maxTimeoutSeconds))
		}
	}
	body := &x402.PaymentRequired{
		X402Version: x402.Version,
		Resource: x402.ResourceInfo{
			Url:         ctx.BaseURL() + ctx.OriginalURL(),
			Description: "Agent inference turn",
			MimeType:    fiber.MIMEApplicationJSON,
		},
		Accepts: accepts,
		Error:   errorCode,
	}
	if encoded, err := x402.EncodeBase64Json(body); err == nil {
		ctx.Set(PaymentRequiredHeader, encoded)
	}
	return ctx.Status(fiber.StatusPaymentRequired).JSON(body)
}

// parseX402Rules keeps x402-method rules and decodes their config, skipping unparseable ones.
func (m *X402AccessMethod) parseX402Rules(rules []models.ExternalPaymentReceiveRule) []x402Rule {
	var out []x402Rule
	for i := range rules {
		if rules[i].Method != methodX402 {
			continue
		}
		config, err := x402.ParseConfig(rules[i].Config)
		if err != nil {
			continue
		}
		out = append(out, x402Rule{rule: &rules[i], config: config})
	}
	return out
}

func (m *X402AccessMethod) matchRules(x402Rules []x402Rule, payload *x402.PaymentPayload) []x402Rule {
	if payload.Scheme != x402.SchemeExact {
		return nil
	}
	if !common.IsHexAddress(payload.RecipientAddress()) {
		return nil
	}
	to := common.HexToAddress(payload.RecipientAddress())

	var candidates []x402Rule
	for _, candidate := range x402Rules {
		if candidate.config.Network == payload.Network && common.HexToAddress(candidate.config.PayTo) == to {
			candidates = append(candidates, candidate)
		}
	}
	return candidates
}

// cheapestEntitledRule picks the lowest-priced pattern-matching rule, skipping unsettleable paid ones.
func (m *X402AccessMethod) cheapestEntitledRule(
	x402Rules []x402Rule,
	payer common.Address,
	facilitators map[string]models.ExternalPaymentFacilitator,
) *x402Rule {
	var best *x402Rule
	var bestPrice *big.Int
	for i := range x402Rules {
		candidate := &x402Rules[i]
		if !m.matchesAddressPatterns(candidate.config.AddressPatterns, payer) {
			continue
		}
		price, ok := new(big.Int).SetString(candidate.rule.Price, 10)
		if !ok {
			continue
		}
		if m.unsettleablePaidRule(candidate, facilitators) {
			continue
		}
		if best == nil || price.Cmp(bestPrice) < 0 {
			best = candidate
			bestPrice = price
		}
	}
	return best
}

// unsettleablePaidRule is a paid rule whose network has no facilitator — never advertised or honored.
func (m *X402AccessMethod) unsettleablePaidRule(
	candidate *x402Rule,
	facilitators map[string]models.ExternalPaymentFacilitator,
) bool {
	_, hasFacilitator := facilitators[candidate.config.Network]
	return !m.isZeroPrice(candidate.rule.Price) && !hasFacilitator
}

// matchesAddressPatterns full-matches the payer's lowercase address; invalid
// patterns never match.
func (m *X402AccessMethod) matchesAddressPatterns(patterns []string, payer common.Address) bool {
	payerHex := strings.ToLower(payer.Hex())
	for _, pattern := range patterns {
		re, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			m.logger.Warn("ignoring uncompilable address pattern",
				zap.String("pattern", pattern),
				zap.Error(err),
			)

			continue
		}
		if re.MatchString(payerHex) {
			return true
		}
	}
	return false
}

func (m *X402AccessMethod) isZeroPrice(price string) bool {
	value, ok := new(big.Int).SetString(price, 10)
	return ok && value.Sign() == 0
}

func (m *X402AccessMethod) coversPrice(payload *x402.PaymentPayload, rule *models.ExternalPaymentReceiveRule) bool {
	value, valueOk := new(big.Int).SetString(payload.Amount(), 10)
	price, priceOk := new(big.Int).SetString(rule.Price, 10)
	return valueOk && priceOk && value.Cmp(price) >= 0
}

func (m *X402AccessMethod) toPaymentRequirements(candidate *x402Rule, maxTimeoutSeconds int) *x402.PaymentRequirements {
	// assetTransferMethod tells the client how to sign; eip3009 also needs the token's EIP-712 domain.
	extra := map[string]any{"assetTransferMethod": candidate.config.AssetTransferMethod}
	if candidate.config.AssetTransferMethod != x402.AssetTransferMethodPermit2 {
		extra["name"] = candidate.config.Eip712Name
		extra["version"] = candidate.config.Eip712Version
	}
	return &x402.PaymentRequirements{
		Scheme:            x402.SchemeExact,
		Network:           candidate.config.Network,
		Amount:            candidate.rule.Price,
		Asset:             candidate.config.Asset,
		PayTo:             candidate.config.PayTo,
		MaxTimeoutSeconds: maxTimeoutSeconds,
		Extra:             extra,
	}
}
