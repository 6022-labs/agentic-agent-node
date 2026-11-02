package requests

import "github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"

// UpdatePaymentSpendPolicyRequest replaces the payees the agent may pay.
type UpdatePaymentSpendPolicyRequest struct {
	PayToPatterns []string `json:"payToPatterns"`
}

func (r *UpdatePaymentSpendPolicyRequest) ToExternalPayload() *models.ExternalUpdatePaymentSpendPolicyPayload {
	return &models.ExternalUpdatePaymentSpendPolicyPayload{PayToPatterns: r.PayToPatterns}
}
