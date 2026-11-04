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
)

// everyoneAddressPattern marks a price as public: only these are advertised in a 402.
const everyoneAddressPattern = ".*"

// x402ReceiveOffer pairs a price with the token it settles in.
type x402ReceiveOffer struct {
	config *models.PaymentReceiveRuleX402Config
	price  *models.PaymentReceiveRulePrice
}

// X402PaymentReceivePolicyEvaluator decides receive access; the gateway only executes the verdict.
type X402PaymentReceivePolicyEvaluator struct {
	paymentReceivePolicyProvider           payment_interfaces.PaymentReceivePolicyProviderInterface
	paymentReceiveRuleRepository           repositories.PaymentReceiveRuleRepositoryInterface
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface
	paymentReceiveRulePriceRepository      repositories.PaymentReceiveRulePriceRepositoryInterface
	x402PaymentVerifier                    payment_interfaces.X402PaymentVerifierInterface
	paymentReceiveSettings                 *runtime_common_settings.PaymentReceiveSettings
}

func NewX402PaymentReceivePolicyEvaluator(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveRuleX402ConfigRepository repositories.PaymentReceiveRuleX402ConfigRepositoryInterface,
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface,
	x402PaymentVerifier payment_interfaces.X402PaymentVerifierInterface,
	paymentReceiveSettings *runtime_common_settings.PaymentReceiveSettings,
) *X402PaymentReceivePolicyEvaluator {
	return &X402PaymentReceivePolicyEvaluator{
		paymentReceivePolicyProvider:           paymentReceivePolicyProvider,
		paymentReceiveRuleRepository:           paymentReceiveRuleRepository,
		paymentReceiveRuleX402ConfigRepository: paymentReceiveRuleX402ConfigRepository,
		paymentReceiveRulePriceRepository:      paymentReceiveRulePriceRepository,
		x402PaymentVerifier:                    x402PaymentVerifier,
		paymentReceiveSettings:                 paymentReceiveSettings,
	}
}

func (e *X402PaymentReceivePolicyEvaluator) Evaluate(
	ctx context.Context,
	credentials *payment_interfaces.PaymentReceiveAccessCredentials,
) (*payment_interfaces.PaymentReceiveAccessDecision, error) {
	offers, err := e.loadOffers(ctx)
	if err != nil {
		return nil, err
	}

	// A JWT-proven caller only ever buys free access; anything priced needs a signature.
	if len(credentials.Address) != 0 && e.matchesFreeOffer(offers, common.HexToAddress(credentials.Address)) {
		return &payment_interfaces.PaymentReceiveAccessDecision{
			Outcome: enums.PaymentReceiveAccessOutcomeFree,
			Payer:   common.HexToAddress(credentials.Address).Hex(),
			Accepts: e.advertisedRequirements(offers),
		}, nil
	}

	if len(credentials.PaymentSignature) == 0 {
		return e.paymentRequired(offers, ""), nil
	}

	payload, err := x402.DecodePaymentPayload(credentials.PaymentSignature)
	if err != nil {
		return e.paymentRequired(offers, "malformed_payment_signature"), nil
	}

	candidates := e.matchConfigs(offers, payload)
	if len(candidates) == 0 {
		return e.paymentRequired(offers, "no_matching_payment_requirements"), nil
	}

	payer, recovered, err := e.recoverAgainstCandidates(payload, candidates)
	if errors.Is(err, payment_interfaces.ErrX402AuthorizationExpired) {
		return e.paymentRequired(offers, "authorization_expired"), nil
	}
	if err != nil {
		return e.paymentRequired(offers, "invalid_payment_signature"), nil
	}

	entitled := e.cheapestEntitledOffer(offers, recovered, payer)
	if entitled == nil {
		return e.paymentRequired(offers, "no_matching_payment_requirements"), nil
	}
	if entitled.price.IsFree() {
		return &payment_interfaces.PaymentReceiveAccessDecision{
			Outcome: enums.PaymentReceiveAccessOutcomeFree,
			Payer:   payer.Hex(),
			Accepts: e.advertisedRequirements(offers),
		}, nil
	}
	if !e.coversPrice(payload, entitled) {
		return e.paymentRequired(offers, "insufficient_amount"), nil
	}

	baseUrl, _ := e.paymentReceiveSettings.BaseUrlFor(entitled.config.Network)
	return &payment_interfaces.PaymentReceiveAccessDecision{
		Outcome: enums.PaymentReceiveAccessOutcomePayable,
		Payer:   payer.Hex(),
		Accepts: e.advertisedRequirements(offers),
		Settlement: &payment_interfaces.PaymentReceiveSettlement{
			Requirements:       *e.toPaymentRequirements(*entitled),
			FacilitatorBaseUrl: baseUrl,
			Network:            entitled.config.Network,
		},
	}, nil
}

// loadOffers lists every x402 price; a pre-mint agent has none.
func (e *X402PaymentReceivePolicyEvaluator) loadOffers(ctx context.Context) ([]x402ReceiveOffer, error) {
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

	offers := []x402ReceiveOffer{}
	for i := range rules {
		if rules[i].Method != enums.PaymentMethodX402 {
			continue
		}
		config, err := e.paymentReceiveRuleX402ConfigRepository.GetByRuleId(ctx, rules[i].Id)
		if err != nil {
			return nil, apperrors.NewInternalError("access_evaluation_failed", err.Error())
		}
		// A rule with no matcher describes nothing, so it entitles nobody.
		if config == nil {
			continue
		}
		prices, err := e.paymentReceiveRulePriceRepository.ListByRuleId(ctx, rules[i].Id)
		if err != nil {
			return nil, apperrors.NewInternalError("access_evaluation_failed", err.Error())
		}
		config.PaymentReceiveRule = rules[i]
		for j := range prices {
			offers = append(offers, x402ReceiveOffer{config: config, price: &prices[j]})
		}
	}
	return offers, nil
}

func (e *X402PaymentReceivePolicyEvaluator) paymentRequired(
	offers []x402ReceiveOffer,
	errorCode string,
) *payment_interfaces.PaymentReceiveAccessDecision {
	return &payment_interfaces.PaymentReceiveAccessDecision{
		Outcome:   enums.PaymentReceiveAccessOutcomePaymentRequired,
		ErrorCode: errorCode,
		Accepts:   e.advertisedRequirements(offers),
	}
}

// advertisedRequirements lists public settleable prices; scoped ones would leak their patterns.
func (e *X402PaymentReceivePolicyEvaluator) advertisedRequirements(
	offers []x402ReceiveOffer,
) []x402.PaymentRequirements {
	accepts := []x402.PaymentRequirements{}
	for i := range offers {
		if !e.isPublic(offers[i].price) || e.unsettleablePaidOffer(offers[i]) {
			continue
		}
		accepts = append(accepts, *e.toPaymentRequirements(offers[i]))
	}
	return accepts
}

func (e *X402PaymentReceivePolicyEvaluator) isPublic(price *models.PaymentReceiveRulePrice) bool {
	for _, pattern := range price.AddressPatterns {
		if pattern == everyoneAddressPattern {
			return true
		}
	}
	return false
}

func (e *X402PaymentReceivePolicyEvaluator) matchesFreeOffer(offers []x402ReceiveOffer, address common.Address) bool {
	for i := range offers {
		if offers[i].price.Matches(address.Hex()) && offers[i].price.IsFree() {
			return true
		}
	}
	return false
}

// matchConfigs returns the tokens the payload could pay, once each.
func (e *X402PaymentReceivePolicyEvaluator) matchConfigs(
	offers []x402ReceiveOffer,
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
	for i := range offers {
		config := offers[i].config
		if config.Network != payload.Network || common.HexToAddress(config.PayTo) != to {
			continue
		}
		if !lo.Contains(candidates, config) {
			candidates = append(candidates, config)
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

// cheapestEntitledOffer picks the lowest matching price of a recovered token, skipping unsettleable paid ones.
func (e *X402PaymentReceivePolicyEvaluator) cheapestEntitledOffer(
	offers []x402ReceiveOffer,
	recovered []*models.PaymentReceiveRuleX402Config,
	payer common.Address,
) *x402ReceiveOffer {
	var best *x402ReceiveOffer
	var bestPrice *big.Int
	for i := range offers {
		candidate := offers[i]
		if !lo.Contains(recovered, candidate.config) || !candidate.price.Matches(payer.Hex()) {
			continue
		}
		price, ok := new(big.Int).SetString(candidate.price.Price, 10)
		if !ok || e.unsettleablePaidOffer(candidate) {
			continue
		}
		if best == nil || price.Cmp(bestPrice) < 0 {
			best = &offers[i]
			bestPrice = price
		}
	}
	return best
}

// unsettleablePaidOffer is a paid price whose network has no facilitator — never advertised or honored.
func (e *X402PaymentReceivePolicyEvaluator) unsettleablePaidOffer(offer x402ReceiveOffer) bool {
	_, hasFacilitator := e.paymentReceiveSettings.BaseUrlFor(offer.config.Network)
	return !offer.price.IsFree() && !hasFacilitator
}

func (e *X402PaymentReceivePolicyEvaluator) coversPrice(payload *x402.PaymentPayload, offer *x402ReceiveOffer) bool {
	// The rule's own method, not whichever field is present.
	signed := payload.AmountFor(offer.config.AssetTransferMethod.String())
	value, valueOk := new(big.Int).SetString(signed, 10)
	price, priceOk := new(big.Int).SetString(offer.price.Price, 10)
	return valueOk && priceOk && value.Cmp(price) >= 0
}

func (e *X402PaymentReceivePolicyEvaluator) toPaymentRequirements(offer x402ReceiveOffer) *x402.PaymentRequirements {
	maxTimeoutSeconds := 0
	if facilitator, ok := (*e.paymentReceiveSettings)[offer.config.Network]; ok {
		maxTimeoutSeconds = facilitator.MaxTimeoutSeconds
	}
	// assetTransferMethod tells the client how to sign; eip3009 also needs the token's EIP-712 domain.
	extra := map[string]any{"assetTransferMethod": offer.config.AssetTransferMethod.String()}
	if offer.config.AssetTransferMethod != enums.AssetTransferMethodPermit2 {
		extra["name"] = lo.FromPtr(offer.config.Eip712Name)
		extra["version"] = lo.FromPtr(offer.config.Eip712Version)
	}
	return &x402.PaymentRequirements{
		Scheme:  x402.SchemeExact,
		Network: offer.config.Network,
		Amount:  offer.price.Price,
		// Stored lowercased; clients get the checksummed form.
		Asset:             common.HexToAddress(offer.config.Asset).Hex(),
		PayTo:             offer.config.PayTo,
		MaxTimeoutSeconds: maxTimeoutSeconds,
		Extra:             extra,
	}
}
