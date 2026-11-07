package use_cases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListPaymentSpendRules struct {
	paymentSpendPolicyProvider       payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendRuleRepository       repositories.PaymentSpendRuleRepositoryInterface
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface
	paymentSpendMethodStrategies     payment_interfaces.PaymentSpendMethodStrategiesInterface
}

func NewListPaymentSpendRules(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
) *ListPaymentSpendRules {
	return &ListPaymentSpendRules{
		paymentSpendPolicyProvider:       paymentSpendPolicyProvider,
		paymentSpendRuleRepository:       paymentSpendRuleRepository,
		paymentSpendRuleWindowRepository: paymentSpendRuleWindowRepository,
		paymentSpendMethodStrategies:     paymentSpendMethodStrategies,
	}
}

func (u *ListPaymentSpendRules) Execute(ctx context.Context) (*responses.ListPaymentSpendRulesResponse, error) {
	policy, err := u.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}
	rules, err := u.paymentSpendRuleRepository.ListByPolicyId(ctx, policy.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}

	// Windows and the token live in their own tables, so join them back per rule.
	out := make([]responses.PaymentSpendRuleResponse, 0, len(rules))
	now := time.Now()
	for i := range rules {
		windows, err := u.paymentSpendRuleWindowRepository.ListByRuleId(ctx, rules[i].Id)
		if err != nil {
			return nil, errors.NewInternalError("list_failed", err)
		}
		var config json.RawMessage
		if strategy, ok := u.paymentSpendMethodStrategies.For(rules[i].Method); ok {
			config, err = strategy.LoadConfig(ctx, rules[i].Id)
			if err != nil {
				return nil, errors.NewInternalError("list_failed", err)
			}
		}
		out = append(out, *responses.ToPaymentSpendRuleResponse(&rules[i], windows, config, now))
	}
	return &responses.ListPaymentSpendRulesResponse{Rules: out}, nil
}
