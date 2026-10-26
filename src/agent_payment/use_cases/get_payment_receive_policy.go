package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type GetPaymentReceivePolicy struct {
	*AbstractPaymentUseCase
}

func NewGetPaymentReceivePolicy(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
	paymentReceiveMethodStrategies payment_interfaces.PaymentReceiveMethodStrategiesInterface,
) *GetPaymentReceivePolicy {
	return &GetPaymentReceivePolicy{
		AbstractPaymentUseCase: NewAbstractPaymentUseCase(
			paymentReceivePolicyProvider,
			paymentReceiveRuleRepository,
			paymentReceiveMethodStrategies,
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
