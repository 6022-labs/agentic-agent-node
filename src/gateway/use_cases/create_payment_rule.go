package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
)

type CreatePaymentRule struct {
	paymentRuleClient gateway_interfaces.PaymentRuleClientInterface
}

func NewCreatePaymentRule(paymentRuleClient gateway_interfaces.PaymentRuleClientInterface) *CreatePaymentRule {
	return &CreatePaymentRule{paymentRuleClient: paymentRuleClient}
}

// Execute forwards as-is: validation is owned by the runtime, and the client
// maps its 400 back to a ValidationError.
func (u *CreatePaymentRule) Execute(ctx context.Context, request *requests.SavePaymentRuleRequest) (*models.ExternalPaymentRule, error) {
	return u.paymentRuleClient.Create(ctx, toExternalPaymentRule(request))
}

func toExternalPaymentRule(request *requests.SavePaymentRuleRequest) *models.ExternalPaymentRule {
	return &models.ExternalPaymentRule{
		Name:              request.Name,
		Network:           request.Network,
		Asset:             request.Asset,
		PayTo:             request.PayTo,
		Price:             request.Price,
		MaxTimeoutSeconds: request.MaxTimeoutSeconds,
		Eip712Name:        request.Eip712Name,
		Eip712Version:     request.Eip712Version,
		AddressPatterns:   request.AddressPatterns,
	}
}
