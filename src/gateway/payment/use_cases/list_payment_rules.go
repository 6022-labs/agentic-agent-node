package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type ListPaymentRules struct {
	paymentRuleClient gateway_interfaces.PaymentRuleClientInterface
}

func NewListPaymentRules(paymentRuleClient gateway_interfaces.PaymentRuleClientInterface) *ListPaymentRules {
	return &ListPaymentRules{paymentRuleClient: paymentRuleClient}
}

func (u *ListPaymentRules) Execute(ctx context.Context) (*responses.ListPaymentRulesResponse, error) {
	rules, err := u.paymentRuleClient.List(ctx)
	if err != nil {
		return nil, err
	}
	return &responses.ListPaymentRulesResponse{Rules: rules}, nil
}
