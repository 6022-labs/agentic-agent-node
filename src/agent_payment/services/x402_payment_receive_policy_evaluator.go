package services

import (
	"context"
	"errors"
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// everyoneAddressPattern marks a rule as public: only these are advertised in a 402.
const everyoneAddressPattern = ".*"

// X402PaymentReceivePolicyEvaluator decides receive access; the gateway only executes the verdict.
type X402PaymentReceivePolicyEvaluator struct {
	logger                                 *zap.Logger
	paymentReceivePolicyProvider           payment_interfaces.PaymentReceivePolicyProviderInterface
	paymentReceiveRuleRepository           repositories.PaymentReceiveRuleRepositoryInterface
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface
	x402PaymentVerifier                    payment_interfaces.X402PaymentVerifierInterface
	paymentReceiveSettings                 *runtime_common_settings.PaymentReceiveSettings
}

func NewX402PaymentReceivePolicyEvaluator(
	logger *zap.Logger,
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface,
	x402PaymentVerifier payment_interfaces.X402PaymentVerifierInterface,
	paymentReceiveSettings *runtime_common_settings.PaymentReceiveSettings,
) *X402PaymentReceivePolicyEvaluator {
	return &X402PaymentReceivePolicyEvaluator{
		logger:                                 logger,
		paymentReceivePolicyProvider:           paymentReceivePolicyProvider,
		paymentReceiveRuleRepository:           paymentReceiveRuleRepository,
		paymentReceiveRuleX402ConfigRepository: paymentReceiveRuleX402ConfigRepository,
		x402PaymentVerifier:                    x402PaymentVerifier,
		paymentReceiveSettings:                 paymentReceiveSettings,
	}
}

func (e *X402PaymentReceivePolicyEvaluator) Evaluate(
	ctx context.Context,
	credentials *payment_interfaces.PaymentReceiveAccessCredentials,
) (*payment_interfaces.PaymentReceiveAccessDecision, error) {
	rules, err := e.loadRules(ctx)
	if err != nil {
		return nil, err
	}

	// A JWT-proven caller only ever buys free access; anything priced needs a signature.
	if len(credentials.Address) != 0 && e.matchesFreeRule(rules, common.HexToAddress(credentials.Address)) {
		return &payment_interfaces.PaymentReceiveAccessDecision{
			Outcome: enums.PaymentReceiveAccessOutcomeFree,
			Payer:   common.HexToAddress(credentials.Address).Hex(),
			Accepts: e.advertisedRequirements(rules),
		}, nil
	}

	if len(credentials.PaymentSignature) == 0 {
		return e.paymentRequired(rules, ""), nil
	}

	payload, err := x402.DecodePaymentPayload(credentials.PaymentSignature)
	if err != nil {
		return e.paymentRequired(rules, "malformed_payment_signature"), nil
	}

	candidates := e.matchRules(rules, payload)
	if len(candidates) == 0 {
		return e.paymentRequired(rules, "no_matching_payment_requirements"), nil
	}

	payer, recovered, err := e.recoverAgainstCandidates(payload, candidates)
	if errors.Is(err, payment_interfaces.ErrX402AuthorizationExpired) {
		return e.paymentRequired(rules, "authorization_expired"), nil
	}
	if err != nil {
		return e.paymentRequired(rules, "invalid_payment_signature"), nil
	}

	entitled := e.cheapestEntitledRule(recovered, payer)
	if entitled == nil {
		return e.paymentRequired(rules, "no_matching_payment_requirements"), nil
	}
	if e.isZeroPrice(entitled.PaymentReceiveRule.Price) {
		return &payment_interfaces.PaymentReceiveAccessDecision{
			Outcome: enums.PaymentReceiveAccessOutcomeFree,
			Payer:   payer.Hex(),
			Accepts: e.advertisedRequirements(rules),
		}, nil
	}
	if !e.coversPrice(payload, entitled) {
		return e.paymentRequired(rules, "insufficient_amount"), nil
	}

	baseUrl, _ := e.paymentReceiveSettings.BaseUrlFor(entitled.Network)
	return &payment_interfaces.PaymentReceiveAccessDecision{
		Outcome: enums.PaymentReceiveAccessOutcomePayable,
		Payer:   payer.Hex(),
		Accepts: e.advertisedRequirements(rules),
		Settlement: &payment_interfaces.PaymentReceiveSettlement{
			Requirements:       *e.toPaymentRequirements(entitled),
			FacilitatorBaseUrl: baseUrl,
			Network:            entitled.Network,
		},
	}, nil
}

// loadRules returns the policy's x402 rules with matchers; a pre-mint agent has none.
func (e *X402PaymentReceivePolicyEvaluator) loadRules(
	ctx context.Context,
) ([]*models.PaymentReceiveRuleX402Config, error) {
	policy, _, err := e.paymentReceivePolicyProvider.Ensure(ctx)
	if err != nil {
		var conflictError *apperrors.ConflictError
		if errors.As(err, &conflictError) {
			return nil, nil
		}
		return nil, err
	}

	rules, err := e.paymentReceiveRuleRepository.ListByPolicyId(ctx, policy.Id)
	if err != nil {
		return nil, apperrors.NewInternalError("access_evaluation_failed", err.Error())
	}

	out := make([]*models.PaymentReceiveRuleX402Config, 0, len(rules))
	for i := range rules {
		if rules[i].Method != enums.PaymentMethodX402 {
			continue
		}
		config, err := e.paymentReceiveRuleX402ConfigRepository.GetByRuleId(ctx, rules[i].Id)
		if err != nil {
			// One unreadable matcher must not lock everyone out; the rule just entitles nobody.
			e.logger.Error("skipping payment receive rule with unreadable x402 config",
				zap.String("ruleId", rules[i].Id.String()), zap.Error(err))
			continue
		}
		// A rule with no matcher describes nothing, so it entitles nobody.
		if config == nil {
			continue
		}
		config.PaymentReceiveRule = rules[i]
		out = append(out, config)
	}
	return out, nil
}

func (e *X402PaymentReceivePolicyEvaluator) paymentRequired(
	rules []*models.PaymentReceiveRuleX402Config,
	errorCode string,
) *payment_interfaces.PaymentReceiveAccessDecision {
	return &payment_interfaces.PaymentReceiveAccessDecision{
		Outcome:   enums.PaymentReceiveAccessOutcomePaymentRequired,
		ErrorCode: errorCode,
		Accepts:   e.advertisedRequirements(rules),
	}
}

// advertisedRequirements lists public settleable rules; scoped ones would leak their patterns.
func (e *X402PaymentReceivePolicyEvaluator) advertisedRequirements(
	rules []*models.PaymentReceiveRuleX402Config,
) []x402.PaymentRequirements {
	accepts := []x402.PaymentRequirements{}
	for i := range rules {
		if !e.isPublic(rules[i]) || e.unsettleablePaidRule(rules[i]) {
			continue
		}
		accepts = append(accepts, *e.toPaymentRequirements(rules[i]))
	}
	return accepts
}

func (e *X402PaymentReceivePolicyEvaluator) isPublic(config *models.PaymentReceiveRuleX402Config) bool {
	for _, pattern := range config.AddressPatterns {
		if pattern == everyoneAddressPattern {
			return true
		}
	}
	return false
}

func (e *X402PaymentReceivePolicyEvaluator) matchesFreeRule(
	rules []*models.PaymentReceiveRuleX402Config,
	address common.Address,
) bool {
	for i := range rules {
		if rules[i].Matches(address.Hex()) && e.isZeroPrice(rules[i].PaymentReceiveRule.Price) {
			return true
		}
	}
	return false
}

func (e *X402PaymentReceivePolicyEvaluator) matchRules(
	rules []*models.PaymentReceiveRuleX402Config,
	payload *x402.PaymentPayload,
) []*models.PaymentReceiveRuleX402Config {
	if payload.Scheme != x402.SchemeExact {
		return nil
	}
	if !common.IsHexAddress(payload.RecipientAddress()) {
		return nil
	}
	to := common.HexToAddress(payload.RecipientAddress())

	var candidates []*models.PaymentReceiveRuleX402Config
	for i := range rules {
		if rules[i].Network == payload.Network && common.HexToAddress(rules[i].PayTo) == to {
			candidates = append(candidates, rules[i])
		}
	}
	return candidates
}

func (e *X402PaymentReceivePolicyEvaluator) recoverAgainstCandidates(
	payload *x402.PaymentPayload,
	candidates []*models.PaymentReceiveRuleX402Config,
) (common.Address, []*models.PaymentReceiveRuleX402Config, error) {
	var payer common.Address
	var recovered []*models.PaymentReceiveRuleX402Config
	var lastErr error
	for i := range candidates {
		address, err := e.x402PaymentVerifier.RecoverPayer(payload, candidates[i])
		if errors.Is(err, payment_interfaces.ErrX402AuthorizationExpired) {
			return common.Address{}, nil, err
		}
		if err != nil {
			lastErr = err
			continue
		}
		payer = address
		recovered = append(recovered, candidates[i])
	}
	if len(recovered) == 0 {
		return common.Address{}, nil, lastErr
	}
	return payer, recovered, nil
}

// cheapestEntitledRule picks the lowest-priced pattern-matching rule, skipping unsettleable paid ones.
func (e *X402PaymentReceivePolicyEvaluator) cheapestEntitledRule(
	rules []*models.PaymentReceiveRuleX402Config,
	payer common.Address,
) *models.PaymentReceiveRuleX402Config {
	var best *models.PaymentReceiveRuleX402Config
	var bestPrice *big.Int
	for i := range rules {
		candidate := rules[i]
		if !candidate.Matches(payer.Hex()) {
			continue
		}
		price, ok := new(big.Int).SetString(candidate.PaymentReceiveRule.Price, 10)
		if !ok || e.unsettleablePaidRule(candidate) {
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
func (e *X402PaymentReceivePolicyEvaluator) unsettleablePaidRule(candidate *models.PaymentReceiveRuleX402Config) bool {
	_, hasFacilitator := e.paymentReceiveSettings.BaseUrlFor(candidate.Network)
	return !e.isZeroPrice(candidate.PaymentReceiveRule.Price) && !hasFacilitator
}

func (e *X402PaymentReceivePolicyEvaluator) isZeroPrice(price string) bool {
	value, ok := new(big.Int).SetString(price, 10)
	return ok && value.Sign() == 0
}

func (e *X402PaymentReceivePolicyEvaluator) coversPrice(
	payload *x402.PaymentPayload,
	candidate *models.PaymentReceiveRuleX402Config,
) bool {
	// The rule's own method, not whichever field is present.
	signed := payload.AmountFor(candidate.AssetTransferMethod.String())
	value, valueOk := new(big.Int).SetString(signed, 10)
	price, priceOk := new(big.Int).SetString(candidate.PaymentReceiveRule.Price, 10)
	return valueOk && priceOk && value.Cmp(price) >= 0
}

func (e *X402PaymentReceivePolicyEvaluator) toPaymentRequirements(
	candidate *models.PaymentReceiveRuleX402Config,
) *x402.PaymentRequirements {
	maxTimeoutSeconds := 0
	if facilitator, ok := (*e.paymentReceiveSettings)[candidate.Network]; ok {
		maxTimeoutSeconds = facilitator.MaxTimeoutSeconds
	}
	// assetTransferMethod tells the client how to sign; eip3009 also needs the token's EIP-712 domain.
	extra := map[string]any{"assetTransferMethod": candidate.AssetTransferMethod.String()}
	if candidate.AssetTransferMethod != enums.AssetTransferMethodPermit2 {
		extra["name"] = lo.FromPtr(candidate.Eip712Name)
		extra["version"] = lo.FromPtr(candidate.Eip712Version)
	}
	return &x402.PaymentRequirements{
		Scheme:            x402.SchemeExact,
		Network:           candidate.Network,
		Amount:            candidate.PaymentReceiveRule.Price,
		Asset:             candidate.Asset,
		PayTo:             candidate.PayTo,
		MaxTimeoutSeconds: maxTimeoutSeconds,
		Extra:             extra,
	}
}
