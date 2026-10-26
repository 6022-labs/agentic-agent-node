package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
)

type DeletePaymentReceiveRule struct {
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface
}

func NewDeletePaymentReceiveRule(
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface,
) *DeletePaymentReceiveRule {
	return &DeletePaymentReceiveRule{paymentReceiveRuleClient: paymentReceiveRuleClient}
}

func (u *DeletePaymentReceiveRule) Execute(
	ctx context.Context,
	request *requests.DeletePaymentReceiveRuleRequest,
) error {
	return u.paymentReceiveRuleClient.Delete(ctx, request.Id.String())
}
