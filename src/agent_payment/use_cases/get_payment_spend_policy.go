package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
)

type GetPaymentSpendPolicy struct {
	*AbstractPaymentSpendUseCase
}

func NewGetPaymentSpendPolicy(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
	paymentSpendMethodStrategies payment_interfaces.PaymentSpendMethodStrategiesInterface,
) *GetPaymentSpendPolicy {
	return &GetPaymentSpendPolicy{
		AbstractPaymentSpendUseCase: NewAbstractPaymentSpendUseCase(
			paymentSpendPolicyProvider,
			paymentSpendRuleRepository,
			paymentSpendMethodStrategies,
		),
	}
}

func (u *GetPaymentSpendPolicy) Execute(ctx context.Context) (*responses.PaymentSpendPolicyResponse, error) {
	policy, err := u.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToPaymentSpendPolicyResponse(policy), nil
}
