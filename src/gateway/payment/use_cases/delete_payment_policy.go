package use_cases

import (
	"context"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
)

type DeletePaymentPolicy struct {
	paymentPolicyClient gateway_interfaces.PaymentPolicyClientInterface
}

func NewDeletePaymentPolicy(paymentPolicyClient gateway_interfaces.PaymentPolicyClientInterface) *DeletePaymentPolicy {
	return &DeletePaymentPolicy{paymentPolicyClient: paymentPolicyClient}
}

func (u *DeletePaymentPolicy) Execute(ctx context.Context) error {
	if err := u.paymentPolicyClient.Delete(ctx); err != nil {
		return apperrors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	return nil
}
