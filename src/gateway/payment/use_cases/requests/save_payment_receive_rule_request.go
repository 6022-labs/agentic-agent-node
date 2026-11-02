package requests

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// SavePaymentReceiveRuleRequest is one token's pricing rule, forwarded as-is; the runtime validates it.
type SavePaymentReceiveRuleRequest struct {
	Name   string                               `json:"name"`
	Method string                               `json:"method,omitempty"`
	Prices []SavePaymentReceiveRulePriceRequest `json:"prices"`
	Config json.RawMessage                      `json:"config"           swaggertype:"object"`
}

// SavePaymentReceiveRulePriceRequest is one price of a rule and the payers it applies to.
type SavePaymentReceiveRulePriceRequest struct {
	Price           string   `json:"price"`
	AddressPatterns []string `json:"addressPatterns"`
}

// ToExternalPaymentReceiveRule maps the request onto the runtime's payment-receive-rule payload.
func (r *SavePaymentReceiveRuleRequest) ToExternalPaymentReceiveRule() *models.ExternalPaymentReceiveRule {
	prices := make([]models.ExternalPaymentReceiveRulePrice, 0, len(r.Prices))
	for _, price := range r.Prices {
		prices = append(prices, models.ExternalPaymentReceiveRulePrice{
			Price:           price.Price,
			AddressPatterns: price.AddressPatterns,
		})
	}
	return &models.ExternalPaymentReceiveRule{
		Name:   r.Name,
		Method: r.Method,
		Prices: prices,
		Config: r.Config,
	}
}
