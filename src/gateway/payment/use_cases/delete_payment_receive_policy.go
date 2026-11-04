package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
)

type DeletePaymentReceivePolicy struct {
	paymentReceivePolicyClient gateway_interfaces.PaymentReceivePolicyClientInterface
}

func NewDeletePaymentReceivePolicy(
	paymentReceivePolicyClient gateway_interfaces.PaymentReceivePolicyClientInterface,
) *DeletePaymentReceivePolicy {
	return &DeletePaymentReceivePolicy{paymentReceivePolicyClient: paymentReceivePolicyClient}
}

func (u *DeletePaymentReceivePolicy) Execute(ctx context.Context) error {
	return u.paymentReceivePolicyClient.Delete(ctx)
}
