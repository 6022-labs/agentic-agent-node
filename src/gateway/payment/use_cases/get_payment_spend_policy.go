package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type GetPaymentSpendPolicy struct {
	paymentSpendPolicyClient gateway_interfaces.PaymentSpendPolicyClientInterface
}

func NewGetPaymentSpendPolicy(
	paymentSpendPolicyClient gateway_interfaces.PaymentSpendPolicyClientInterface,
) *GetPaymentSpendPolicy {
	return &GetPaymentSpendPolicy{paymentSpendPolicyClient: paymentSpendPolicyClient}
}

func (u *GetPaymentSpendPolicy) Execute(ctx context.Context) (*responses.GetPaymentSpendPolicyResponse, error) {
	policy, err := u.paymentSpendPolicyClient.Get(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToGetPaymentSpendPolicyResponse(policy), nil
}
