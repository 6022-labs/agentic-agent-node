package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// PaymentSpendRuleClientInterface is the gateway-side port for the runtime's
// spend-rule endpoints. List returns the effective rules in table order.
type PaymentSpendRuleClientInterface interface {
	List(ctx context.Context) ([]models.ExternalPaymentSpendRule, error)
	Create(ctx context.Context, rule *models.ExternalPaymentSpendRule) (*models.ExternalPaymentSpendRule, error)
	Update(
		ctx context.Context,
		id string,
		rule *models.ExternalPaymentSpendRule,
	) (*models.ExternalPaymentSpendRule, error)
	Delete(ctx context.Context, id string) error
}
