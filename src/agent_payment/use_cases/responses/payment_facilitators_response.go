package responses

import runtime_common_settings "github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"

// PaymentFacilitatorResponse is one network's x402 facilitator.
type PaymentFacilitatorResponse struct {
	Network           string `json:"network"`
	BaseUrl           string `json:"baseUrl"`
	MaxTimeoutSeconds int    `json:"maxTimeoutSeconds"`
}

// PaymentFacilitatorsResponse is the configured facilitators served to the
// gateway's payment gate. Internal only — BaseUrl may carry credentials, so it
// is never proxied to the browser.
type PaymentFacilitatorsResponse struct {
	Facilitators []PaymentFacilitatorResponse `json:"facilitators"`
}

func ToPaymentFacilitatorsResponse(entries []runtime_common_settings.PaymentFacilitator) *PaymentFacilitatorsResponse {
	facilitators := make([]PaymentFacilitatorResponse, len(entries))
	for i, entry := range entries {
		facilitators[i] = PaymentFacilitatorResponse{
			Network:           entry.Network,
			BaseUrl:           entry.BaseUrl,
			MaxTimeoutSeconds: entry.MaxTimeoutSeconds,
		}
	}
	return &PaymentFacilitatorsResponse{Facilitators: facilitators}
}
