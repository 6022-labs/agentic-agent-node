package services

import (
	"context"
	"math/big"

	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
	"github.com/google/uuid"
)

// X402PaymentSpendPolicyEvaluator denies by default: a quote no rule matches is refused.
type X402PaymentSpendPolicyEvaluator struct {
	paymentSpendPolicyProvider           payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendRuleRepository           repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface
	unitOfWorkFactory                    runtime_common_interfaces.UnitOfWorkFactoryInterface
}

func NewX402PaymentSpendPolicyEvaluator(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleX402ConfigRepository repositories.PaymentSpendRuleX402ConfigRepositoryInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *X402PaymentSpendPolicyEvaluator {
	return &X402PaymentSpendPolicyEvaluator{
		paymentSpendPolicyProvider:           paymentSpendPolicyProvider,
		paymentSpendRuleRepository:           paymentSpendRuleRepository,
		paymentSpendRuleX402ConfigRepository: paymentSpendRuleX402ConfigRepository,
		unitOfWorkFactory:                    unitOfWorkFactory,
	}
}

func (e *X402PaymentSpendPolicyEvaluator) Authorize(
	ctx context.Context,
	quote *payment_interfaces.X402PaymentQuote,
	amount *big.Int,
) (bool, error) {
	policy, err := e.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return false, err
	}

	// Deciding and recording share one transaction, so concurrent signatures cannot overrun a window.
	unitOfWork, err := e.unitOfWorkFactory.Start(ctx)
	if err != nil {
		return false, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	defer func() { _ = unitOfWork.Rollback() }()

	authorized, err := e.recordSpend(unitOfWork.Context(), policy.Id, quote, amount)
	if err != nil {
		return false, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	// A refusal wrote nothing, so there is nothing worth keeping.
	if !authorized {
		return false, nil
	}
	if err := unitOfWork.Commit(); err != nil {
		return false, errors.NewInternalError("spend_policy_failed", err.Error())
	}
	return true, nil
}

// recordSpend records the spend against every matching rule, reporting false when any of them refuses.
func (e *X402PaymentSpendPolicyEvaluator) recordSpend(
	ctx context.Context,
	policyId uuid.UUID,
	quote *payment_interfaces.X402PaymentQuote,
	amount *big.Int,
) (bool, error) {
	rules, err := e.paymentSpendRuleRepository.ListByPolicyIdForUpdate(ctx, policyId)
	if err != nil {
		return false, err
	}

	matching, err := e.matchingRules(ctx, rules, quote)
	if err != nil {
		return false, err
	}
	// No match is a refusal, not a fallthrough: an unmatched quote is one the owner never described.
	if len(matching) == 0 {
		return false, nil
	}

	now := time.Now()
	for i := range matching {
		covered, err := matching[i].CoversPayment(amount)
		if err != nil {
			return false, err
		}
		recorded, err := matching[i].RecordWindowSpend(amount, now)
		if err != nil {
			return false, err
		}
		if !covered || !recorded {
			return false, nil
		}
	}

	// Recording against every match keeps each rule's budget its own, not just the binding one.
	for i := range matching {
		if _, err := e.paymentSpendRuleRepository.Update(ctx, &matching[i]); err != nil {
			return false, err
		}
	}
	return true, nil
}

// matchingRules returns the x402 rules this quote falls under.
func (e *X402PaymentSpendPolicyEvaluator) matchingRules(
	ctx context.Context,
	rules []models.PaymentSpendRule,
	quote *payment_interfaces.X402PaymentQuote,
) ([]models.PaymentSpendRule, error) {
	var matching []models.PaymentSpendRule
	for i := range rules {
		if rules[i].Method != enums.PaymentMethodX402 {
			continue
		}
		config, err := e.paymentSpendRuleX402ConfigRepository.GetByRuleId(ctx, rules[i].Id)
		if err != nil {
			return nil, errors.NewInternalError("spend_policy_failed", err.Error())
		}
		// A rule with no matcher describes nothing, so it grants nothing.
		if config == nil {
			continue
		}
		if !config.Matches(quote.Network, quote.Asset, quote.PayTo, quote.AssetTransferMethod.String()) {
			continue
		}
		matching = append(matching, rules[i])
	}
	return matching, nil
}
