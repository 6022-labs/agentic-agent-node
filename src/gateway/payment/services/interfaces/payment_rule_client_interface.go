package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// PaymentRuleClientInterface is the gateway-side port for the runtime's
// pricing-rule endpoints. List returns the effective rules in table order.
type PaymentRuleClientInterface interface {
	List(ctx context.Context) ([]models.ExternalPaymentRule, error)
	Create(ctx context.Context, rule *models.ExternalPaymentRule) (*models.ExternalPaymentRule, error)
	Update(ctx context.Context, id string, rule *models.ExternalPaymentRule) (*models.ExternalPaymentRule, error)
	Delete(ctx context.Context, id string) error
}
