package requests

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// SavePaymentSpendRuleRequest is one spend rule in the gateway's payment-spend-policy API.
type SavePaymentSpendRuleRequest struct {
	Name      string          `json:"name"`
	MaxAmount string          `json:"maxAmount"`
	Method    string          `json:"method,omitempty"`
	Config    json.RawMessage `json:"config"           swaggertype:"object"`
}

// ToExternalPaymentSpendRule maps the request onto the runtime's spend-rule payload.
func (r *SavePaymentSpendRuleRequest) ToExternalPaymentSpendRule() *models.ExternalPaymentSpendRule {
	return &models.ExternalPaymentSpendRule{
		Name:      r.Name,
		MaxAmount: r.MaxAmount,
		Method:    r.Method,
		Config:    r.Config,
	}
}
