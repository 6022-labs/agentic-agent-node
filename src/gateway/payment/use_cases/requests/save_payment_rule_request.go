package requests

import "encoding/json"

// SavePaymentRuleRequest is one pricing rule in the gateway's payment-policy API.
// Validation is owned by the runtime — the gateway forwards as-is. Name/Price/
// Method are method-neutral; Config carries the method-specific settings.
type SavePaymentRuleRequest struct {
	Name   string          `json:"name"`
	Price  string          `json:"price"`
	Method string          `json:"method,omitempty"`
	Config json.RawMessage `json:"config" swaggertype:"object"`
}
