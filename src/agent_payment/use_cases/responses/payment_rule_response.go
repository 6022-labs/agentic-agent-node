package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentRuleResponse is the wire shape of a pricing rule. The facilitator is
// network-level config, not rule data — it's served by GET /internal/config/payment-facilitators.
type PaymentRuleResponse struct {
	Id                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Network           string    `json:"network"`
	Asset             string    `json:"asset"`
	PayTo             string    `json:"payTo"`
	Price             string    `json:"price"`
	MaxTimeoutSeconds int       `json:"maxTimeoutSeconds"`
	Eip712Name        string    `json:"eip712Name"`
	Eip712Version     string    `json:"eip712Version"`
	AddressPatterns   []string  `json:"addressPatterns"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func ToPaymentRuleResponse(m *models.PaymentRule) *PaymentRuleResponse {
	if m == nil {
		return nil
	}
	return &PaymentRuleResponse{
		Id:                m.Id,
		Name:              m.Name,
		Network:           m.Network,
		Asset:             m.Asset,
		PayTo:             m.PayTo,
		Price:             m.Price,
		MaxTimeoutSeconds: m.MaxTimeoutSeconds,
		Eip712Name:        m.Eip712Name,
		Eip712Version:     m.Eip712Version,
		AddressPatterns:   m.AddressPatterns,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func ToPaymentRuleResponses(rules []models.PaymentRule) []PaymentRuleResponse {
	out := make([]PaymentRuleResponse, len(rules))
	for i := range rules {
		out[i] = *ToPaymentRuleResponse(&rules[i])
	}
	return out
}
