package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendRuleResponse is the method-neutral wire shape of a spend rule.
type PaymentSpendRuleResponse struct {
	Id                  uuid.UUID                        `json:"id"`
	Name                string                           `json:"name"`
	MaxAmountPerPayment *string                          `json:"maxAmountPerPayment,omitempty"`
	Windows             []PaymentSpendRuleWindowResponse `json:"windows"`
	Method              string                           `json:"method"`
	Config              json.RawMessage                  `json:"config"                        swaggertype:"object"`
	CreatedAt           time.Time                        `json:"createdAt"`
	UpdatedAt           time.Time                        `json:"updatedAt"`
}

// PaymentSpendRuleWindowResponse reports one window's budget and its spend as of now.
type PaymentSpendRuleWindowResponse struct {
	WindowSeconds int        `json:"windowSeconds"`
	MaxAmount     string     `json:"maxAmount"`
	Spent         string     `json:"spent"`
	ResetsAt      *time.Time `json:"resetsAt,omitempty"`
}

// ToPaymentSpendRuleResponse joins the rule with its windows and method config, reporting periods as of now.
func ToPaymentSpendRuleResponse(
	m *models.PaymentSpendRule,
	windows []models.PaymentSpendRuleWindow,
	config json.RawMessage,
	now time.Time,
) *PaymentSpendRuleResponse {
	if m == nil {
		return nil
	}
	out := make([]PaymentSpendRuleWindowResponse, 0, len(windows))
	for i := range windows {
		out = append(out, PaymentSpendRuleWindowResponse{
			WindowSeconds: windows[i].WindowSeconds,
			MaxAmount:     windows[i].MaxAmount,
			Spent:         windows[i].SpentAt(now),
			ResetsAt:      windows[i].ResetsAt(now),
		})
	}
	return &PaymentSpendRuleResponse{
		Id:                  m.Id,
		Name:                m.Name,
		MaxAmountPerPayment: m.MaxAmountPerPayment,
		Windows:             out,
		Method:              m.Method.String(),
		Config:              config,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
