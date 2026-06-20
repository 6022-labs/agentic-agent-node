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

type ListPaymentRules struct {
	*AbstractPaymentUseCase
}

func NewListPaymentRules(
	agentRepository identity_repositories.AgentRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *ListPaymentRules {
	return &ListPaymentRules{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			agentRepository,
			paymentPolicyRepository,
			paymentRuleRepository,
			paymentMethodStrategies,
			token6022Settings,
		),
	}
}

func (u *ListPaymentRules) Execute(ctx context.Context) (*responses.ListPaymentRulesResponse, error) {
	policy, _, err := u.ensurePaymentPolicy(ctx)
	if err != nil {
		return nil, err
	}
	rules, err := u.paymentRuleRepository.ListByPolicyId(ctx, policy.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	// The config lives in each method's own table, so join it back per rule.
	out := make([]responses.PaymentRuleResponse, 0, len(rules))
	for i := range rules {
		var config json.RawMessage
		if strategy, ok := u.paymentMethodStrategies.For(rules[i].Method); ok {
			config, err = strategy.LoadConfig(ctx, rules[i].Id)
			if err != nil {
				return nil, errors.NewInternalError("list_failed", err.Error())
			}
		}
		out = append(out, *responses.ToPaymentRuleResponse(&rules[i], config))
	}
	return &responses.ListPaymentRulesResponse{Rules: out}, nil
}
