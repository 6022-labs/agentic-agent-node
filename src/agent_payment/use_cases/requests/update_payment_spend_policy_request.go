package requests

import payment_helpers "github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"

// UpdatePaymentSpendPolicyRequest replaces who the agent may pay; an empty list pays nobody.
type UpdatePaymentSpendPolicyRequest struct {
	PayToPatterns []string `json:"payToPatterns"`
}

func (r *UpdatePaymentSpendPolicyRequest) ValidateAndSanitize() error {
	sanitized, err := payment_helpers.SanitizeAddressPatterns("payToPatterns", r.PayToPatterns)
	if err != nil {
		return err
	}
	r.PayToPatterns = sanitized
	return nil
}
