package use_cases

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListPaymentSpendRules struct {
	*AbstractPaymentSpendUseCase
}

func NewListPaymentSpendRules(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
) *ListPaymentSpendRules {
	return &ListPaymentSpendRules{
		AbstractPaymentSpendUseCase: NewAbstractPaymentSpendUseCase(
			paymentSpendPolicyProvider,
			paymentSpendRuleRepository,
			paymentSpendMethodStrategies,
		),
	}
}

func (u *ListPaymentSpendRules) Execute(ctx context.Context) (*responses.ListPaymentSpendRulesResponse, error) {
	policy, err := u.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}
	rules, err := u.paymentSpendRuleRepository.ListByPolicyId(ctx, policy.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	// The matcher lives in each method's own table, so join it back per rule.
	out := make([]responses.PaymentSpendRuleResponse, 0, len(rules))
	for i := range rules {
		var config json.RawMessage
		if strategy, ok := u.paymentSpendMethodStrategies.For(rules[i].Method); ok {
			config, err = strategy.LoadConfig(ctx, rules[i].Id)
			if err != nil {
				return nil, errors.NewInternalError("list_failed", err.Error())
			}
		}
		out = append(out, *responses.ToPaymentSpendRuleResponse(&rules[i], config))
	}
	return &responses.ListPaymentSpendRulesResponse{Rules: out}, nil
}
