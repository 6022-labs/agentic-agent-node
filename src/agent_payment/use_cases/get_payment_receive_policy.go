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

type GetPaymentReceivePolicy struct {
	*AbstractPaymentUseCase
}

func NewGetPaymentReceivePolicy(
	agentRepository identity_repositories.AgentRepositoryInterface,
	agentAddressRepository identity_repositories.AgentAddressRepositoryInterface,
	paymentReceivePolicyRepository repositories.PaymentReceivePolicyRepositoryInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentMethodStrategies payment_interfaces.PaymentMethodStrategiesInterface,
	token6022Settings *runtime_common_settings.Token6022Settings,
) *GetPaymentReceivePolicy {
	return &GetPaymentReceivePolicy{
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

func (u *GetPaymentReceivePolicy) Execute(ctx context.Context) (*responses.PaymentReceivePolicyResponse, error) {
	policy, agent, err := u.ensurePaymentReceivePolicy(ctx)
	if err != nil {
		return nil, err
	}
	evmAddress, err := u.evmAddress(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("policy_failed", err.Error())
	}
	return responses.ToPaymentReceivePolicyResponse(policy, evmAddress), nil
}
