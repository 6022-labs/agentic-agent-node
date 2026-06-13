package use_cases

import (
	"context"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
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
		return nil, apperrors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	return &responses.ListPaymentRulesResponse{Rules: rules}, nil
}
