package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type UpdatePaymentSpendPolicy struct {
	paymentSpendPolicyClient gateway_interfaces.PaymentSpendPolicyClientInterface
}

func NewUpdatePaymentSpendPolicy(
	paymentSpendPolicyClient gateway_interfaces.PaymentSpendPolicyClientInterface,
) *UpdatePaymentSpendPolicy {
	return &UpdatePaymentSpendPolicy{paymentSpendPolicyClient: paymentSpendPolicyClient}
}

func (u *UpdatePaymentSpendPolicy) Execute(
	ctx context.Context,
	request *requests.UpdatePaymentSpendPolicyRequest,
) (*responses.GetPaymentSpendPolicyResponse, error) {
	policy, err := u.paymentSpendPolicyClient.Update(ctx, request.ToExternalPayload())
	if err != nil {
		return nil, err
	}
	return responses.ToGetPaymentSpendPolicyResponse(policy), nil
}
