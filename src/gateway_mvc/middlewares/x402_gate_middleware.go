package middlewares

import (
	"errors"
	"math/big"
	"regexp"
	"slices"
	"strings"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/mvc"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models/x402"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

const (
	PaymentRequiredHeader  = "PAYMENT-REQUIRED"
	PaymentSignatureHeader = "PAYMENT-SIGNATURE"
	PaymentResponseHeader  = "PAYMENT-RESPONSE"

	X402PayerLocalKey = "x402.payer"

	// Rules carrying this pattern are public and advertised in 402s.
	everyoneAddressPattern = ".*"
)

// X402GateMiddleware executes the payment policy over x402 v2: the payer gets
// the cheapest rule whose patterns match them (price "0" served free, never
// settled), no match = refused.
type X402GateMiddleware struct {
	logger                   *zap.Logger
	paymentRuleClient        gateway_interfaces.PaymentRuleClientInterface
	paymentFacilitatorClient gateway_interfaces.PaymentFacilitatorClientInterface
	x402PaymentVerifier      gateway_interfaces.X402PaymentVerifierInterface
	x402FacilitatorClient    gateway_interfaces.X402FacilitatorClientInterface
}

func NewX402GateMiddleware(
	logger *zap.Logger,
	paymentRuleClient gateway_interfaces.PaymentRuleClientInterface,
	paymentFacilitatorClient gateway_interfaces.PaymentFacilitatorClientInterface,
	x402PaymentVerifier gateway_interfaces.X402PaymentVerifierInterface,
	x402FacilitatorClient gateway_interfaces.X402FacilitatorClientInterface,
) *X402GateMiddleware {
	return &X402GateMiddleware{
		logger:                   logger,
		paymentRuleClient:        paymentRuleClient,
		paymentFacilitatorClient: paymentFacilitatorClient,
		x402PaymentVerifier:      x402PaymentVerifier,
		x402FacilitatorClient:    x402FacilitatorClient,
	}
}

func (m *X402GateMiddleware) Handle(ctx *fiber.Ctx) error {
	rules, err := m.paymentRuleClient.List(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, apperrors.NewUnavailableError("payment_rules_unavailable", err.Error()))
	}
	// The facilitator is network-level config, joined here: a rule is settleable
	// only if its network has one.
	facilitators, err := m.paymentFacilitatorClient.Get(ctx.UserContext())
	if err != nil {
		return mvc.WriteError(ctx, apperrors.NewUnavailableError("payment_facilitators_unavailable", err.Error()))
	}

	header := ctx.Get(PaymentSignatureHeader)
	if len(header) == 0 {
		return m.paymentRequired(ctx, rules, facilitators, "")
	}

	payload, err := x402.DecodePaymentPayload(header)
	if err != nil {
		return m.paymentRequired(ctx, rules, facilitators, "malformed_payment_signature")
	}

	candidates := matchRules(rules, payload)
	if len(candidates) == 0 {
		return m.paymentRequired(ctx, rules, facilitators, "no_matching_payment_requirements")
	}

	// Candidates share network+payTo but may differ in asset (= EIP-712
	// domain), so recovery is attempted per candidate.
	payer, recovered, err := m.recoverAgainstCandidates(payload, candidates)
	if errors.Is(err, gateway_interfaces.ErrX402AuthorizationExpired) {
		return m.paymentRequired(ctx, rules, facilitators, "authorization_expired")
	}
	if err != nil {
		return m.paymentRequired(ctx, rules, facilitators, "invalid_payment_signature")
	}

	// The payer takes the identity role the JWT fills on other routes.
	ctx.Locals(X402PayerLocalKey, payer.Hex())
	ctx.Locals(AuthAddressLocalKey, payer.Hex())

	rule := cheapestEntitledRule(recovered, payer, facilitators)
	if rule == nil {
		return m.paymentRequired(ctx, rules, facilitators, "no_matching_payment_requirements")
	}

	if isZeroPrice(rule.Price) {
		return ctx.Next()
	}
	if !coversPrice(payload, rule) {
		return m.paymentRequired(ctx, rules, facilitators, "insufficient_amount")
	}

	facilitatorBaseUrl := facilitators[rule.Network]
	requirements := toPaymentRequirements(rule)
	verifyResponse, err := m.x402FacilitatorClient.Verify(ctx.UserContext(), facilitatorBaseUrl, payload, requirements)
	if err != nil {
		return mvc.WriteError(ctx, apperrors.NewUnavailableError("facilitator_unavailable", err.Error()))
	}
	if !verifyResponse.IsValid {
		return m.paymentRequired(ctx, rules, facilitators, verifyResponse.InvalidReason)
	}

	if err := ctx.Next(); err != nil {
		return err
	}
	if ctx.Response().StatusCode() >= fiber.StatusBadRequest {
		return nil
	}

	settleResponse, err := m.x402FacilitatorClient.Settle(ctx.UserContext(), facilitatorBaseUrl, payload, requirements)
	if err != nil || !settleResponse.Success {
		if err != nil {
			m.logger.Error("x402 settlement failed", zap.Error(err))
		} else {
			m.logger.Error("x402 settlement rejected", zap.String("reason", settleResponse.ErrorReason))
		}
		// The caller wasn't charged, so they must not receive the response.
		ctx.Response().ResetBody()
		return m.paymentRequired(ctx, rules, facilitators, "settlement_failed")
	}

	encoded, err := x402.EncodeBase64Json(settleResponse)
	if err != nil {
		m.logger.Error("failed to encode settle response", zap.Error(err))
		return nil
	}
	ctx.Set(PaymentResponseHeader, encoded)
	return nil
}

func (m *X402GateMiddleware) recoverAgainstCandidates(payload *x402.PaymentPayload, candidates []*models.ExternalPaymentRule) (common.Address, []*models.ExternalPaymentRule, error) {
	var payer common.Address
	var recovered []*models.ExternalPaymentRule
	var lastErr error
	for _, rule := range candidates {
		p, err := m.x402PaymentVerifier.RecoverPayer(payload, rule)
		if errors.Is(err, gateway_interfaces.ErrX402AuthorizationExpired) {
			return common.Address{}, nil, err
		}
		if err != nil {
			lastErr = err
			continue
		}
		payer = p
		recovered = append(recovered, rule)
	}
	if len(recovered) == 0 {
		return common.Address{}, nil, lastErr
	}
	return payer, recovered, nil
}

// paymentRequired writes the spec 402 (header + body), advertising only the
// public rules — listing scoped ones would leak their patterns.
func (m *X402GateMiddleware) paymentRequired(ctx *fiber.Ctx, rules []models.ExternalPaymentRule, facilitators map[string]string, errorCode string) error {
	accepts := []x402.PaymentRequirements{}
	for i := range rules {
		if slices.Contains(rules[i].AddressPatterns, everyoneAddressPattern) && !unsettleablePaidRule(&rules[i], facilitators) {
			accepts = append(accepts, *toPaymentRequirements(&rules[i]))
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

func matchRules(rules []models.ExternalPaymentRule, payload *x402.PaymentPayload) []*models.ExternalPaymentRule {
	if payload.Scheme != x402.SchemeExact {
		return nil
	}
	if !common.IsHexAddress(payload.Payload.Authorization.To) {
		return nil
	}
	to := common.HexToAddress(payload.Payload.Authorization.To)

	var candidates []*models.ExternalPaymentRule
	for i := range rules {
		rule := &rules[i]
		if rule.Network == payload.Network && common.HexToAddress(rule.PayTo) == to {
			candidates = append(candidates, rule)
		}
	}
	return candidates
}

// cheapestEntitledRule picks the lowest-priced rule whose patterns match the
// payer, skipping paid rules whose network has no facilitator (unsettleable).
func cheapestEntitledRule(rules []*models.ExternalPaymentRule, payer common.Address, facilitators map[string]string) *models.ExternalPaymentRule {
	var best *models.ExternalPaymentRule
	var bestPrice *big.Int
	for _, rule := range rules {
		if !matchesAddressPatterns(rule.AddressPatterns, payer) {
			continue
		}
		price, ok := new(big.Int).SetString(rule.Price, 10)
		if !ok {
			continue
		}
		if unsettleablePaidRule(rule, facilitators) {
			continue
		}
		if best == nil || price.Cmp(bestPrice) < 0 {
			best = rule
			bestPrice = price
		}
	}
	return best
}

// unsettleablePaidRule is true for a paid rule whose network has no configured
// facilitator — it can't be verified or settled, so it's never advertised or
// honored.
func unsettleablePaidRule(rule *models.ExternalPaymentRule, facilitators map[string]string) bool {
	_, hasFacilitator := facilitators[rule.Network]
	return !isZeroPrice(rule.Price) && !hasFacilitator
}

// matchesAddressPatterns full-matches the payer's lowercase address; invalid
// patterns never match.
func matchesAddressPatterns(patterns []string, payer common.Address) bool {
	payerHex := strings.ToLower(payer.Hex())
	for _, pattern := range patterns {
		re, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			continue
		}
		if re.MatchString(payerHex) {
			return true
		}
	}
	return false
}

func isZeroPrice(price string) bool {
	value, ok := new(big.Int).SetString(price, 10)
	return ok && value.Sign() == 0
}

func coversPrice(payload *x402.PaymentPayload, rule *models.ExternalPaymentRule) bool {
	value, valueOk := new(big.Int).SetString(payload.Payload.Authorization.Value, 10)
	price, priceOk := new(big.Int).SetString(rule.Price, 10)
	return valueOk && priceOk && value.Cmp(price) >= 0
}

func toPaymentRequirements(rule *models.ExternalPaymentRule) *x402.PaymentRequirements {
	return &x402.PaymentRequirements{
		Scheme:            x402.SchemeExact,
		Network:           rule.Network,
		Amount:            rule.Price,
		Asset:             rule.Asset,
		PayTo:             rule.PayTo,
		MaxTimeoutSeconds: rule.MaxTimeoutSeconds,
		Extra: map[string]any{
			"name":    rule.Eip712Name,
			"version": rule.Eip712Version,
		},
	}
}
