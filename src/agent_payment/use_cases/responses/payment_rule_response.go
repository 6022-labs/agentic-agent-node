package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentRuleResponse is the method-neutral wire shape of a pricing rule.
// Config carries the method-specific settings as raw JSON.
type PaymentRuleResponse struct {
	Id        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Price     string          `json:"price"`
	Method    string          `json:"method"`
	Config    json.RawMessage `json:"config" swaggertype:"object"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// ToPaymentRuleResponse joins the rule with its method config blob (loaded by
// the method's strategy), since the config lives in its own table now.
func ToPaymentRuleResponse(m *models.PaymentRule, config json.RawMessage) *PaymentRuleResponse {
	if m == nil {
		return nil
	}
	return &PaymentRuleResponse{
		Id:        m.Id,
		Name:      m.Name,
		Price:     m.Price,
		Method:    m.Method.String(),
		Config:    config,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
