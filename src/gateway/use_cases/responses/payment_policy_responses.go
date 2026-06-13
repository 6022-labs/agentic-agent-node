package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
)

// GetPaymentPolicyResponse is the policy head returned by /config/payment-policy.
type GetPaymentPolicyResponse struct {
	Id                 string    `json:"id"`
	AgentId            string    `json:"agentId"`
	AgentWalletAddress string    `json:"agentWalletAddress"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func ToGetPaymentPolicyResponse(p *models.ExternalPaymentPolicyPayload) *GetPaymentPolicyResponse {
	if p == nil {
		return nil
	}
	return &GetPaymentPolicyResponse{
		Id:                 p.Id,
		AgentId:            p.AgentId,
		AgentWalletAddress: p.AgentWalletAddress,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

// ListPaymentRulesResponse is the wire shape of /config/payment-policy/rules.
// Rules carry no facilitator URL, so the model is returned directly.
type ListPaymentRulesResponse struct {
	Rules []models.ExternalPaymentRule `json:"rules"`
}
