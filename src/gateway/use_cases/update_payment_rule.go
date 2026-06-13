package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
)

type UpdatePaymentRule struct {
	paymentRuleClient gateway_interfaces.PaymentRuleClientInterface
}

func NewUpdatePaymentRule(paymentRuleClient gateway_interfaces.PaymentRuleClientInterface) *UpdatePaymentRule {
	return &UpdatePaymentRule{paymentRuleClient: paymentRuleClient}
}

func (u *UpdatePaymentRule) Execute(ctx context.Context, id string, request *requests.SavePaymentRuleRequest) (*models.ExternalPaymentRule, error) {
	return u.paymentRuleClient.Update(ctx, id, toExternalPaymentRule(request))
}
