package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentReceiveRuleResponse is the method-neutral wire shape of a token's pricing rule.
type PaymentReceiveRuleResponse struct {
	Id        uuid.UUID                         `json:"id"`
	Name      string                            `json:"name"`
	Method    string                            `json:"method"`
	Prices    []PaymentReceiveRulePriceResponse `json:"prices"`
	Config    json.RawMessage                   `json:"config"    swaggertype:"object"`
	CreatedAt time.Time                         `json:"createdAt"`
	UpdatedAt time.Time                         `json:"updatedAt"`
}

// PaymentReceiveRulePriceResponse is one price of a rule and the payers it applies to.
type PaymentReceiveRulePriceResponse struct {
	Price           string   `json:"price"`
	AddressPatterns []string `json:"addressPatterns"`
}

// ToPaymentReceiveRuleResponse joins the rule with its prices and its method config blob.
func ToPaymentReceiveRuleResponse(
	m *models.PaymentReceiveRule,
	prices []models.PaymentReceiveRulePrice,
	config json.RawMessage,
) *PaymentReceiveRuleResponse {
	if m == nil {
		return nil
	}
	out := make([]PaymentReceiveRulePriceResponse, 0, len(prices))
	for i := range prices {
		out = append(out, PaymentReceiveRulePriceResponse{
			Price:           prices[i].Price,
			AddressPatterns: prices[i].AddressPatterns,
		})
	}
	return &PaymentReceiveRuleResponse{
		Id:        m.Id,
		Name:      m.Name,
		Method:    m.Method.String(),
		Prices:    out,
		Config:    config,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
