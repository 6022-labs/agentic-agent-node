package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/responses"
)

type CreatePaymentReceiveRule struct {
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface
}

func NewCreatePaymentReceiveRule(
	paymentReceiveRuleClient gateway_interfaces.PaymentReceiveRuleClientInterface,
) *CreatePaymentReceiveRule {
	return &CreatePaymentReceiveRule{paymentReceiveRuleClient: paymentReceiveRuleClient}
}

// Execute forwards as-is: validation is owned by the runtime, and the client
// maps its 400 back to a ValidationError.
func (u *CreatePaymentReceiveRule) Execute(
	ctx context.Context,
	request *requests.SavePaymentReceiveRuleRequest,
) (*responses.PaymentReceiveRuleResponse, error) {
	rule, err := u.paymentReceiveRuleClient.Create(ctx, request.ToExternalPaymentReceiveRule())
	if err != nil {
		return nil, err
	}
	return responses.ToPaymentReceiveRuleResponse(rule), nil
}
