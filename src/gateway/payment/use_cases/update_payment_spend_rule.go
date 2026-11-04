package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type UpdatePaymentSpendRule struct {
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface
}

func NewUpdatePaymentSpendRule(
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface,
) *UpdatePaymentSpendRule {
	return &UpdatePaymentSpendRule{paymentSpendRuleClient: paymentSpendRuleClient}
}

func (u *UpdatePaymentSpendRule) Execute(
	ctx context.Context,
	request *requests.UpdatePaymentSpendRuleRequest,
) (*responses.PaymentSpendRuleResponse, error) {
	rule, err := u.paymentSpendRuleClient.Update(ctx, request.Id.String(), request.ToExternalPaymentSpendRule())
	if err != nil {
		return nil, err
	}
	return responses.ToPaymentSpendRuleResponse(rule), nil
}
