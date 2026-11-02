package use_cases

import (
	"context"

	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

type GetPaymentSpendPolicy struct {
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface
}

func NewGetPaymentSpendPolicy(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
) *GetPaymentSpendPolicy {
	return &GetPaymentSpendPolicy{paymentSpendPolicyProvider: paymentSpendPolicyProvider}
}

func (u *GetPaymentSpendPolicy) Execute(ctx context.Context) (*responses.PaymentSpendPolicyResponse, error) {
	policy, err := u.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToPaymentSpendPolicyResponse(policy), nil
}
