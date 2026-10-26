package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// PaymentReceiveRuleClientInterface is the gateway-side port for the runtime's
// pricing-rule endpoints. List returns the effective rules in table order.
type PaymentReceiveRuleClientInterface interface {
	List(ctx context.Context) ([]models.ExternalPaymentReceiveRule, error)
	Create(ctx context.Context, rule *models.ExternalPaymentReceiveRule) (*models.ExternalPaymentReceiveRule, error)
	Update(
		ctx context.Context,
		id string,
		rule *models.ExternalPaymentReceiveRule,
	) (*models.ExternalPaymentReceiveRule, error)
	Delete(ctx context.Context, id string) error
}
