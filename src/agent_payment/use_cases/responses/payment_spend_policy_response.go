package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentSpendPolicyResponse is the policy head returned by /internal/config/payment-spend-policy.
type PaymentSpendPolicyResponse struct {
	Id            uuid.UUID `json:"id"`
	AgentId       uuid.UUID `json:"agentId"`
	PayToPatterns []string  `json:"payToPatterns"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func ToPaymentSpendPolicyResponse(p *models.PaymentSpendPolicy) *PaymentSpendPolicyResponse {
	if p == nil {
		return nil
	}
	return &PaymentSpendPolicyResponse{
		Id:            p.Id,
		AgentId:       p.AgentId,
		PayToPatterns: p.PayToPatterns,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
