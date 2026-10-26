package responses

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

// PaymentReceivePolicyResponse is the policy head returned by
// /internal/config/payment-receive-policy; rules are served by their own endpoint.
// AgentWalletAddress lets the console pre-fill rules paying the agent.
type PaymentReceivePolicyResponse struct {
	Id                 uuid.UUID `json:"id"`
	AgentId            uuid.UUID `json:"agentId"`
	AgentWalletAddress string    `json:"agentWalletAddress"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func ToPaymentReceivePolicyResponse(
	p *models.PaymentReceivePolicy,
	agentWalletAddress string,
) *PaymentReceivePolicyResponse {
	if p == nil {
		return nil
	}
	return &PaymentReceivePolicyResponse{
		Id:                 p.Id,
		AgentId:            p.AgentId,
		AgentWalletAddress: agentWalletAddress,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}
