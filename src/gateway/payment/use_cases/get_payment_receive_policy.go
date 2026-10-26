package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type GetPaymentReceivePolicy struct {
	paymentReceivePolicyClient gateway_interfaces.PaymentReceivePolicyClientInterface
}

func NewGetPaymentReceivePolicy(
	paymentReceivePolicyClient gateway_interfaces.PaymentReceivePolicyClientInterface,
) *GetPaymentReceivePolicy {
	return &GetPaymentReceivePolicy{paymentReceivePolicyClient: paymentReceivePolicyClient}
}

func (u *GetPaymentReceivePolicy) Execute(ctx context.Context) (*responses.GetPaymentReceivePolicyResponse, error) {
	payload, err := u.paymentReceivePolicyClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToGetPaymentReceivePolicyResponse(payload), nil
}
