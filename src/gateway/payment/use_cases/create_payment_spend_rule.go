package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type CreatePaymentSpendRule struct {
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface
}

func NewCreatePaymentSpendRule(
	paymentSpendRuleClient gateway_interfaces.PaymentSpendRuleClientInterface,
) *CreatePaymentSpendRule {
	return &CreatePaymentSpendRule{paymentSpendRuleClient: paymentSpendRuleClient}
}

func (u *CreatePaymentSpendRule) Execute(
	ctx context.Context,
	request *requests.SavePaymentSpendRuleRequest,
) (*responses.PaymentSpendRuleResponse, error) {
	rule, err := u.paymentSpendRuleClient.Create(ctx, request.ToExternalPaymentSpendRule())
	if err != nil {
		return nil, err
	}
	return responses.ToPaymentSpendRuleResponse(rule), nil
}
