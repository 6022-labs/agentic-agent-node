package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type ListPaymentReceiveRules struct {
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface
}

func NewListPaymentReceiveRules(
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface,
) *ListPaymentReceiveRules {
	return &ListPaymentReceiveRules{paymentReceiveRuleClient: paymentReceiveRuleClient}
}

func (u *ListPaymentReceiveRules) Execute(ctx context.Context) (*responses.ListPaymentReceiveRulesResponse, error) {
	rules, err := u.paymentReceiveRuleClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToListPaymentReceiveRulesResponse(rules), nil
}
