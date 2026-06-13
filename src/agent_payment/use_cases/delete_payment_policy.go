package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentPolicy struct {
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface
}

func NewDeletePaymentPolicy(
	paymentPolicyRepository repositories.PaymentPolicyRepositoryInterface,
) *DeletePaymentPolicy {
	return &DeletePaymentPolicy{paymentPolicyRepository: paymentPolicyRepository}
}

func (u *DeletePaymentPolicy) Execute(ctx context.Context) error {
	if err := u.paymentPolicyRepository.Delete(ctx); err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	return nil
}
