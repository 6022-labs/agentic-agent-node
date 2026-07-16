package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
)

type DeletePaymentRule struct {
	paymentRuleClient gateway_interfaces.PaymentRuleClientInterface
}

func NewDeletePaymentRule(paymentRuleClient gateway_interfaces.PaymentRuleClientInterface) *DeletePaymentRule {
	return &DeletePaymentRule{paymentRuleClient: paymentRuleClient}
}

func (u *DeletePaymentRule) Execute(ctx context.Context, request *requests.DeletePaymentRuleRequest) error {
	return u.paymentRuleClient.Delete(ctx, request.Id.String())
}
