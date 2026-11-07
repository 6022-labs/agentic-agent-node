package use_cases

import (
	"context"

	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type GetPaymentReceivePolicy struct {
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface
}

func NewGetPaymentReceivePolicy(
	paymentReceivePolicyProvider payment_interfaces.PaymentReceivePolicyProviderInterface,
) *GetPaymentReceivePolicy {
	return &GetPaymentReceivePolicy{paymentReceivePolicyProvider: paymentReceivePolicyProvider}
}

func (u *GetPaymentReceivePolicy) Execute(ctx context.Context) (*responses.PaymentReceivePolicyResponse, error) {
	policy, agent, err := u.paymentReceivePolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}
	evmAddress, err := u.paymentReceivePolicyProvider.EvmAddress(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("policy_failed", err)
	}
	return responses.ToPaymentReceivePolicyResponse(policy, evmAddress), nil
}
