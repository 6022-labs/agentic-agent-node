package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type ListPaymentSpendRules struct {
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface
}

func NewListPaymentSpendRules(
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface,
) *ListPaymentSpendRules {
	return &ListPaymentSpendRules{paymentSpendRuleClient: paymentSpendRuleClient}
}

func (u *ListPaymentSpendRules) Execute(ctx context.Context) (*responses.ListPaymentSpendRulesResponse, error) {
	rules, err := u.paymentSpendRuleClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return responses.ToListPaymentSpendRulesResponse(rules), nil
}
