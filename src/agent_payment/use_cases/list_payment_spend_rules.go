package use_cases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
)

type ListPaymentSpendRules struct {
	*AbstractPaymentSpendUseCase
}

func NewListPaymentSpendRules(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendRuleWindowRepository repositories.PaymentSpendRuleWindowRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *ListPaymentSpendRules {
	return &ListPaymentSpendRules{
		AbstractPaymentSpendUseCase: NewAbstractPaymentSpendUseCase(
			paymentSpendPolicyProvider,
			paymentSpendRuleRepository,
			paymentSpendRuleWindowRepository,
			paymentSpendMethodStrategies,
			unitOfWorkFactory,
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

	// Windows and the token live in their own tables, so join them back per rule.
	out := make([]responses.PaymentSpendRuleResponse, 0, len(rules))
	now := time.Now()
	for i := range rules {
		windows, err := u.paymentSpendRuleWindowRepository.ListByRuleId(ctx, rules[i].Id)
		if err != nil {
			return nil, errors.NewInternalError("list_failed", err.Error())
		}
		var config json.RawMessage
		if strategy, ok := u.paymentSpendMethodStrategies.For(rules[i].Method); ok {
			config, err = strategy.LoadConfig(ctx, rules[i].Id)
			if err != nil {
				return nil, errors.NewInternalError("list_failed", err.Error())
			}
		}
		out = append(out, *responses.ToPaymentSpendRuleResponse(&rules[i], windows, config, now))
	}
	return &responses.ListPaymentSpendRulesResponse{Rules: out}, nil
}
