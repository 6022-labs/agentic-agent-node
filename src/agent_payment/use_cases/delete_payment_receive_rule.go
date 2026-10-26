package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentReceiveRule struct {
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface
}

func NewDeletePaymentReceiveRule(
	paymentReceiveRuleRepository repositories.PaymentReceiveRuleRepositoryInterface,
) *DeletePaymentReceiveRule {
	return &DeletePaymentReceiveRule{paymentReceiveRuleRepository: paymentReceiveRuleRepository}
}

func (u *DeletePaymentReceiveRule) Execute(
	ctx context.Context,
	request *requests.DeletePaymentReceiveRuleRequest,
) error {
	rule, err := u.paymentReceiveRuleRepository.GetById(ctx, request.Id)
	if err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	if rule == nil {
		return errors.NewNotFoundError("rule_not_found", "")
	}
	if err := u.paymentReceiveRuleRepository.DeleteById(ctx, request.Id); err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	return nil
}
