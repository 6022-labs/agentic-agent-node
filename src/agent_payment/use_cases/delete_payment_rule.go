package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeletePaymentRule struct {
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface
}

func NewDeletePaymentRule(
	paymentRuleRepository repositories.PaymentRuleRepositoryInterface,
) *DeletePaymentRule {
	return &DeletePaymentRule{paymentRuleRepository: paymentRuleRepository}
}

func (u *DeletePaymentRule) Execute(ctx context.Context, request *requests.DeletePaymentRuleRequest) error {
	rule, err := u.paymentRuleRepository.GetById(ctx, request.Id)
	if err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	if rule == nil {
		return errors.NewNotFoundError("rule_not_found", "")
	}
	if err := u.paymentRuleRepository.DeleteById(ctx, request.Id); err != nil {
		return errors.NewInternalError("delete_failed", err.Error())
	}
	return nil
}
