package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type UpdatePaymentSpendPolicy struct {
	paymentSpendPolicyProvider   payment_interfaces.PaymentSpendPolicyProviderInterface
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface
}

func NewUpdatePaymentSpendPolicy(
	paymentSpendPolicyProvider payment_interfaces.PaymentSpendPolicyProviderInterface,
	paymentSpendPolicyRepository repositories.PaymentSpendPolicyRepositoryInterface,
) *UpdatePaymentSpendPolicy {
	return &UpdatePaymentSpendPolicy{
		paymentSpendPolicyProvider:   paymentSpendPolicyProvider,
		paymentSpendPolicyRepository: paymentSpendPolicyRepository,
	}
}

func (u *UpdatePaymentSpendPolicy) Execute(
	ctx context.Context,
	request *requests.UpdatePaymentSpendPolicyRequest,
) (*responses.PaymentSpendPolicyResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	policy, err := u.paymentSpendPolicyProvider.Ensure(ctx)
	if err != nil {
		return nil, err
	}

	policy.PayToPatterns = request.PayToPatterns
	updated, err := u.paymentSpendPolicyRepository.Update(ctx, policy)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	return responses.ToPaymentSpendPolicyResponse(updated), nil
}
