package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type GetPaymentPolicy struct {
	paymentPolicyClient gateway_interfaces.PaymentPolicyClientInterface
}

func NewGetPaymentPolicy(paymentPolicyClient gateway_interfaces.PaymentPolicyClientInterface) *GetPaymentPolicy {
	return &GetPaymentPolicy{paymentPolicyClient: paymentPolicyClient}
}

func (u *GetPaymentPolicy) Execute(ctx context.Context) (*responses.GetPaymentPolicyResponse, error) {
	payload, err := u.paymentPolicyClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToGetPaymentPolicyResponse(payload), nil
}
