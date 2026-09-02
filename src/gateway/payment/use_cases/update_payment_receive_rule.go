package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type UpdatePaymentReceiveRule struct {
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface
}

func NewUpdatePaymentReceiveRule(
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface,
) *UpdatePaymentReceiveRule {
	return &UpdatePaymentReceiveRule{paymentReceiveRuleClient: paymentReceiveRuleClient}
}

func (u *UpdatePaymentReceiveRule) Execute(
	ctx context.Context,
	request *requests.UpdatePaymentReceiveRuleRequest,
) (*responses.PaymentReceiveRuleResponse, error) {
	rule, err := u.paymentReceiveRuleClient.Update(ctx, request.Id.String(), request.ToExternalPaymentReceiveRule())
	if err != nil {
		return nil, err
	}
	return responses.ToPaymentReceiveRuleResponse(rule), nil
}
