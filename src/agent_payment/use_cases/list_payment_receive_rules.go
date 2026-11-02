package use_cases

import (
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_common/services/interfaces"
)

type ListPaymentReceiveRules struct {
	*AbstractPaymentUseCase
}

func NewListPaymentReceiveRules(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveRulePriceRepository repositories.PaymentReceiveRulePriceRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
	unitOfWorkFactory runtime_common_interfaces.UnitOfWorkFactoryInterface,
) *ListPaymentReceiveRules {
	return &ListPaymentReceiveRules{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			paymentReceivePolicyProvider,
			paymentReceiveRuleRepository,
			paymentReceiveRulePriceRepository,
			paymentReceiveMethodStrategies,
			unitOfWorkFactory,
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

	out := make([]responses.PaymentReceiveRuleResponse, 0, len(rules))
	for i := range rules {
		prices, err := u.paymentReceiveRulePriceRepository.ListByRuleId(ctx, rules[i].Id)
		if err != nil {
			return nil, errors.NewInternalError("list_failed", err.Error())
		}
		var config json.RawMessage
		if strategy, ok := u.paymentReceiveMethodStrategies.For(rules[i].Method); ok {
			config, err = strategy.LoadConfig(ctx, rules[i].Id)
			if err != nil {
				return nil, errors.NewInternalError("list_failed", err.Error())
			}
		}
		out = append(out, *responses.ToPaymentReceiveRuleResponse(&rules[i], prices, config))
	}
	return &responses.ListPaymentReceiveRulesResponse{Rules: out}, nil
}
