package use_cases

import (
	"context"
	"encoding/json"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

type ListPaymentReceiveRules struct {
	*AbstractPaymentUseCase
}

func NewListPaymentReceiveRules(
	agentRepository identity_repositories.AgentRepositoryInterface,
	agentAddressRepository identity_repositories.AgentAddressRepositoryInterface,
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *ListPaymentReceiveRules {
	return &ListPaymentReceiveRules{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			agentRepository,
			agentAddressRepository,
			paymentReceivePolicyRepository,
			paymentReceiveRuleRepository,
			paymentMethodStrategies,
			token6022Settings,
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
		if strategy, ok := u.paymentMethodStrategies.For(rules[i].Method); ok {
			config, err = strategy.LoadConfig(ctx, rules[i].Id)
			if err != nil {
				return nil, errors.NewInternalError("list_failed", err.Error())
			}
		}
		out = append(out, *responses.ToPaymentReceiveRuleResponse(&rules[i], config))
	}
	return &responses.ListPaymentReceiveRulesResponse{Rules: out}, nil
}
