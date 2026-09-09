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
	Method              string          `json:"method"`
	Config              json.RawMessage `json:"config"              swaggertype:"object"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

// ToPaymentSpendRuleResponse joins the rule with its method config blob (loaded by
// the method's strategy), since the config lives in its own table now.
func ToPaymentSpendRuleResponse(m *models.PaymentSpendRule, config json.RawMessage) *PaymentSpendRuleResponse {
	if m == nil {
		return nil
	}
	return &PaymentSpendRuleResponse{
		Id:                  m.Id,
		Name:                m.Name,
		MaxAmountPerPayment: m.MaxAmountPerPayment,
		Method:              m.Method.String(),
		Config:              config,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
