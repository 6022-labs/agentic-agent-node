package use_cases

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListPaymentReceiveRules struct {
	*AbstractPaymentUseCase
}

func NewListPaymentReceiveRules(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
) *ListPaymentReceiveRules {
	return &ListPaymentReceiveRules{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			paymentReceivePolicyProvider,
			paymentReceiveRuleRepository,
			paymentReceiveMethodStrategies,
		),
	}
}

func (u *ListPaymentReceiveRules) Execute(ctx context.Context) (*responses.ListPaymentReceiveRulesResponse, error) {
	policy, _, err := u.ensurePaymentReceivePolicy(ctx)
	if err != nil {
		return nil, err
	}
	rules, err := u.paymentReceiveRuleRepository.ListByPolicyId(ctx, policy.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	// The config lives in each method's own table, so join it back per rule.
	out := make([]responses.PaymentReceiveRuleResponse, 0, len(rules))
	for i := range rules {
		var config json.RawMessage
		if strategy, ok := u.paymentReceiveMethodStrategies.For(rules[i].Method); ok {
			config, err = strategy.LoadConfig(ctx, rules[i].Id)
			if err != nil {
				return nil, errors.NewInternalError("list_failed", err.Error())
			}
		}
		out = append(out, *responses.ToPaymentReceiveRuleResponse(&rules[i], config))
	}
	return &responses.ListPaymentReceiveRulesResponse{Rules: out}, nil
}
