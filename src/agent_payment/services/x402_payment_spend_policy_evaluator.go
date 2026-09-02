package services

import (
	"context"
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// X402PaymentSpendPolicyEvaluator denies by default: a quote no rule matches is refused.
type X402PaymentSpendPolicyEvaluator struct {
	paymentSpendPolicyProvider           payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendRuleRepository           repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface
}

func NewX402PaymentSpendPolicyEvaluator(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface,
) *X402PaymentSpendPolicyEvaluator {
	return &X402PaymentSpendPolicyEvaluator{
		paymentSpendPolicyProvider:           paymentSpendPolicyProvider,
		paymentSpendRuleRepository:           paymentSpendRuleRepository,
		paymentSpendRuleX402ConfigRepository: paymentSpendRuleX402ConfigRepository,
	}
}

func (e *X402PaymentSpendPolicyEvaluator) IsAuthorized(
	ctx context.Context,
	quote *payment_interfaces.X402PaymentQuote,
	amount *big.Int,
) (bool, error) {
	policy, err := e.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return false, err
	}
	rules, err := e.paymentSpendRuleRepository.ListByPolicyId(ctx, policy.Id)
	if err != nil {
		return false, errors.NewInternalError("spend_policy_failed", err.Error())
	}

	lowestCap, err := e.effectiveCap(ctx, rules, quote)
	if err != nil {
		return false, err
	}
	// No match is a refusal, not a fallthrough: an unmatched quote is one the owner never described.
	if lowestCap == nil {
		return false, nil
	}
	return amount.Cmp(lowestCap) <= 0, nil
}

// effectiveCap is the smallest cap among the rules that match; nil means nothing matched.
func (e *X402PaymentSpendPolicyEvaluator) effectiveCap(
	ctx context.Context,
	rules []models.PaymentSpendRule,
	quote *payment_interfaces.X402PaymentQuote,
) (*big.Int, error) {
	var lowestCap *big.Int
	for i := range rules {
		if rules[i].Method != enums.PaymentMethodX402 {
			continue
		}
		config, err := e.paymentSpendRuleX402ConfigRepository.GetByRuleId(ctx, rules[i].Id)
		if err != nil {
			return nil, errors.NewInternalError("spend_policy_failed", err.Error())
		}
		// A rule whose matcher never landed cannot describe anything, so it grants nothing.
		if config == nil {
			continue
		}
		if !config.Matches(quote.Network, quote.Asset, quote.PayTo, quote.AssetTransferMethod.String()) {
			continue
		}
		maxAmount, ok := new(big.Int).SetString(rules[i].MaxAmount, 10)
		if !ok || maxAmount.Sign() < 0 {
			return nil, errors.NewInternalError(
				"spend_policy_failed",
				"rule "+rules[i].Id.String()+" has an unusable cap",
			)
		}
		if lowestCap == nil || maxAmount.Cmp(lowestCap) < 0 {
			lowestCap = maxAmount
		}
	}
	return lowestCap, nil
}
