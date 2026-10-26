package requests

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// SavePaymentReceiveRuleRequest is one pricing rule in the gateway's payment-receive-policy API.
// Validation is owned by the runtime — the gateway forwards as-is. Name/Price/
// Method are method-neutral; Config carries the method-specific settings.
type SavePaymentReceiveRuleRequest struct {
	Name   string          `json:"name"`
	Price  string          `json:"price"`
	Method string          `json:"method,omitempty"`
	Config json.RawMessage `json:"config"           swaggertype:"object"`
}

// ToExternalPaymentReceiveRule maps the request onto the runtime's payment-receive-rule payload.
func (r *SavePaymentReceiveRuleRequest) ToExternalPaymentReceiveRule() *models.ExternalPaymentReceiveRule {
	return &models.ExternalPaymentReceiveRule{
		Name:   r.Name,
		Price:  r.Price,
		Method: r.Method,
		Config: r.Config,
	}
}
