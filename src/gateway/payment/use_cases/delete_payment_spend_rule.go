package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
)

type DeletePaymentSpendRule struct {
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface
}

func NewDeletePaymentSpendRule(
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface,
) *DeletePaymentSpendRule {
	return &DeletePaymentSpendRule{paymentSpendRuleClient: paymentSpendRuleClient}
}

func (u *DeletePaymentSpendRule) Execute(ctx context.Context, request *requests.DeletePaymentSpendRuleRequest) error {
	return u.paymentSpendRuleClient.Delete(ctx, request.Id.String())
}
