package requests

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// SavePaymentRuleRequest is one pricing rule in the gateway's payment-policy API.
// Validation is owned by the runtime — the gateway forwards as-is. Name/Price/
// Method are method-neutral; Config carries the method-specific settings.
type SavePaymentRuleRequest struct {
	Name   string          `json:"name"`
	Price  string          `json:"price"`
	Method string          `json:"method,omitempty"`
	Config json.RawMessage `json:"config" swaggertype:"object"`
}

// ToExternalPaymentRule maps the request onto the runtime's payment-rule payload.
func (r *SavePaymentRuleRequest) ToExternalPaymentRule() *models.ExternalPaymentRule {
	return &models.ExternalPaymentRule{
		Name:   r.Name,
		Price:  r.Price,
		Method: r.Method,
		Config: r.Config,
	}
}
