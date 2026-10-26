package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// GetPaymentReceivePolicyResponse is the policy head returned by /config/payment-receive-policy.
type GetPaymentReceivePolicyResponse struct {
	Id                 string    `json:"id"`
	AgentId            string    `json:"agentId"`
	AgentWalletAddress string    `json:"agentWalletAddress"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func ToGetPaymentReceivePolicyResponse(p *models.ExternalPaymentReceivePolicyPayload) *GetPaymentReceivePolicyResponse {
	if p == nil {
		return nil
	}
	return &GetPaymentReceivePolicyResponse{
		Id:                 p.Id,
		AgentId:            p.AgentId,
		AgentWalletAddress: p.AgentWalletAddress,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

// ListPaymentReceiveRulesResponse is the wire shape of /config/payment-receive-policy/rules.
// Rules carry no facilitator URL, so the model is returned directly.
type ListPaymentReceiveRulesResponse struct {
	Rules []models.ExternalPaymentReceiveRule `json:"rules"`
}
