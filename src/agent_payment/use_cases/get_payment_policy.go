package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
)

type GetPaymentPolicy struct {
	*AbstractPaymentUseCase
}

func NewGetPaymentPolicy(
	agentRepository identity_repositories.AgentRepositoryInterface,
	agentAddressRepository identity_repositories.AgentAddressRepositoryInterface,
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *GetPaymentPolicy {
	return &GetPaymentPolicy{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			agentRepository,
			agentAddressRepository,
			paymentPolicyRepository,
			paymentRuleRepository,
			paymentMethodStrategies,
			token6022Settings,
		),
	}
}

func (u *GetPaymentPolicy) Execute(ctx context.Context) (*responses.PaymentPolicyResponse, error) {
	policy, agent, err := u.ensurePaymentPolicy(ctx)
	if err != nil {
		return nil, err
	}
	evmAddress, err := u.evmAddress(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("policy_failed", err.Error())
	}
	return responses.ToPaymentPolicyResponse(policy, evmAddress), nil
}
