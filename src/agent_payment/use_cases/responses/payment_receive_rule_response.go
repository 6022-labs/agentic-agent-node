package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentReceiveRuleResponse is the method-neutral wire shape of a pricing rule.
// Config carries the method-specific settings as raw JSON.
type PaymentReceiveRuleResponse struct {
	Id        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Price     string          `json:"price"`
	Method    string          `json:"method"`
	Config    json.RawMessage `json:"config"    swaggertype:"object"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

// ToPaymentReceiveRuleResponse joins the rule with its method config blob (loaded by
// the method's strategy), since the config lives in its own table now.
func ToPaymentReceiveRuleResponse(m *models.PaymentReceiveRule, config json.RawMessage) *PaymentReceiveRuleResponse {
	if m == nil {
		return nil
	}
	return &PaymentReceiveRuleResponse{
		Id:        m.Id,
		Name:      m.Name,
		Price:     m.Price,
		Method:    m.Method.String(),
		Config:    config,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
