package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentSpendRule struct {
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface
}

func NewDeletePaymentSpendRule(
	paymentSpendRuleRepository repositories.PaymentSpendRuleRepositoryInterface,
) *DeletePaymentSpendRule {
	return &DeletePaymentSpendRule{paymentSpendRuleRepository: paymentSpendRuleRepository}
}

func (u *DeletePaymentSpendRule) Execute(
	ctx context.Context,
	request *requests.DeletePaymentSpendRuleRequest,
) error {
	rule, err := u.paymentSpendRuleRepository.GetById(ctx, request.Id)
	if err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	if rule == nil {
		return errors.NewNotFoundError("rule_not_found", "")
	}
	if err := u.paymentSpendRuleRepository.DeleteById(ctx, request.Id); err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	return nil
}
