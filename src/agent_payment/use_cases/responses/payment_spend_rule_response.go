package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendRuleResponse is the method-neutral wire shape of a spend rule.
type PaymentSpendRuleResponse struct {
	Id                  uuid.UUID       `json:"id"`
	Name                string          `json:"name"`
	MaxAmountPerPayment string          `json:"maxAmountPerPayment"`
	WindowSeconds       int             `json:"windowSeconds"`
	MaxAmountPerWindow  string          `json:"maxAmountPerWindow"`
	WindowSpent         string          `json:"windowSpent"`
	WindowResetsAt      *time.Time      `json:"windowResetsAt,omitempty"`
	Method              string          `json:"method"`
	Config              json.RawMessage `json:"config"                   swaggertype:"object"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

// ToPaymentSpendRuleResponse joins the rule with its method config and reports the period as of now.
func ToPaymentSpendRuleResponse(
	m *models.PaymentSpendRule,
	config json.RawMessage,
	now time.Time,
) *PaymentSpendRuleResponse {
	if m == nil {
		return nil
	}
	return &PaymentSpendRuleResponse{
		Id:                  m.Id,
		Name:                m.Name,
		MaxAmountPerPayment: m.MaxAmountPerPayment,
		WindowSeconds:       m.WindowSeconds,
		MaxAmountPerWindow:  m.MaxAmountPerWindow,
		WindowSpent:         m.SpentInWindow(now),
		WindowResetsAt:      m.WindowResetsAt(now),
		Method:              m.Method.String(),
		Config:              config,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
