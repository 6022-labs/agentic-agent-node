package requests

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// SavePaymentSpendRuleRequest is one spend rule in the gateway's payment-spend-policy API.
type SavePaymentSpendRuleRequest struct {
	Name                string                              `json:"name"`
	MaxAmountPerPayment *string                             `json:"maxAmountPerPayment,omitempty"`
	Windows             []SavePaymentSpendRuleWindowRequest `json:"windows"`
	Method              string                              `json:"method,omitempty"`
	Config              json.RawMessage                     `json:"config"                        swaggertype:"object"`
}

// SavePaymentSpendRuleWindowRequest is one recurring budget of a spend rule.
type SavePaymentSpendRuleWindowRequest struct {
	WindowSeconds int    `json:"windowSeconds"`
	MaxAmount     string `json:"maxAmount"`
}

// ToExternalPaymentSpendRule maps the request onto the runtime's spend-rule payload.
func (r *SavePaymentSpendRuleRequest) ToExternalPaymentSpendRule() *models.ExternalPaymentSpendRule {
	windows := make([]models.ExternalPaymentSpendRuleWindow, 0, len(r.Windows))
	for _, window := range r.Windows {
		windows = append(windows, models.ExternalPaymentSpendRuleWindow{
			WindowSeconds: window.WindowSeconds,
			MaxAmount:     window.MaxAmount,
		})
	}
	return &models.ExternalPaymentSpendRule{
		Name:                r.Name,
		MaxAmountPerPayment: r.MaxAmountPerPayment,
		Windows:             windows,
		Method:              r.Method,
		Config:              r.Config,
	}
}
