package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
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
) (*models.ExternalPaymentReceiveRule, error) {
	return u.paymentReceiveRuleClient.Update(ctx, request.Id.String(), request.ToExternalPaymentReceiveRule())
}
