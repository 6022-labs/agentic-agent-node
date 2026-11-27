package responses

// Statuses a call to a peer agent can end in; a quoted 402 is an outcome, not a failure.
const (
	CallAgentStatusOk              = "ok"
	CallAgentStatusPaymentRequired = "payment_required"
)

// CallAgentResponse is one answered turn, or the quote the peer wants paid first.
type CallAgentResponse struct {
	Status          string                   `json:"status"`
	EndpointUrl     string                   `json:"endpointUrl"`
	AgentName       string                   `json:"agentName,omitempty"`
	Text            string                   `json:"text,omitempty"`
	ContextId       string                   `json:"contextId,omitempty"`
	PaymentRequired *PaymentRequiredResponse `json:"paymentRequired,omitempty"`
}

// PaymentRequiredResponse is the peer's x402 quote; one accepts entry feeds the signing tool directly.
type PaymentRequiredResponse struct {
	X402Version int                          `json:"x402Version"`
	Resource    string                       `json:"resource,omitempty"`
	Error       string                       `json:"error,omitempty"`
	Accepts     []PaymentRequirementResponse `json:"accepts"`
}

type PaymentRequirementResponse struct {
	Scheme            string         `json:"scheme"`
	Network           string         `json:"network"`
	Amount            string         `json:"amount"`
	Asset             string         `json:"asset"`
	PayTo             string         `json:"payTo"`
	MaxTimeoutSeconds int            `json:"maxTimeoutSeconds"`
	Extra             map[string]any `json:"extra,omitempty"`
}
