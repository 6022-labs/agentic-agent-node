package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
)

type DeletePaymentSpendPolicy struct {
	paymentSpendPolicyClient gateway_interfaces.PaymentSpendPolicyClientInterface
}

func NewDeletePaymentSpendPolicy(
	paymentSpendPolicyClient gateway_interfaces.PaymentSpendPolicyClientInterface,
) *DeletePaymentSpendPolicy {
	return &DeletePaymentSpendPolicy{paymentSpendPolicyClient: paymentSpendPolicyClient}
}

func (u *DeletePaymentSpendPolicy) Execute(ctx context.Context) error {
	return u.paymentSpendPolicyClient.Delete(ctx)
}
