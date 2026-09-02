package responses

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	payment_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_payment/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
)

// AuthorizePaymentReceiveAccessResponse is the runtime's verdict; the gateway only executes it.
type AuthorizePaymentReceiveAccessResponse struct {
	Outcome enums.PaymentReceiveAccessOutcome `json:"outcome"`
	// Payer is empty when the outcome refuses.
	Payer string `json:"payer,omitempty"`
	// ErrorCode lands in the x402 body's `error` field.
	ErrorCode string `json:"errorCode,omitempty"`
	// Accepts carries public rules only; scoped ones would leak their patterns.
	Accepts []x402.PaymentRequirements `json:"accepts"`
	// Settlement is set only when the outcome is payable.
	Settlement *PaymentReceiveSettlementResponse `json:"settlement,omitempty"`
}

// PaymentReceiveSettlementResponse pairs the matched requirements with the facilitator that can settle them.
type PaymentReceiveSettlementResponse struct {
	Requirements       x402.PaymentRequirements `json:"requirements"`
	FacilitatorBaseUrl string                   `json:"facilitatorBaseUrl"`
	Network            string                   `json:"network"`
}

func ToAuthorizePaymentReceiveAccessResponse(
	decision *payment_interfaces.PaymentReceiveAccessDecision,
) *AuthorizePaymentReceiveAccessResponse {
	response := &AuthorizePaymentReceiveAccessResponse{
		Outcome:   decision.Outcome,
		Payer:     decision.Payer,
		ErrorCode: decision.ErrorCode,
		Accepts:   decision.Accepts,
	}
	if decision.Settlement != nil {
		response.Settlement = &PaymentReceiveSettlementResponse{
			Requirements:       decision.Settlement.Requirements,
			FacilitatorBaseUrl: decision.Settlement.FacilitatorBaseUrl,
			Network:            decision.Settlement.Network,
		}
	}
	return response
}
