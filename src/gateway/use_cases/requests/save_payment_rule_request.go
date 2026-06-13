package requests

// SavePaymentRuleRequest is one pricing rule in the gateway's payment-policy API.
// Validation is owned by the runtime — the gateway forwards as-is.
type SavePaymentRuleRequest struct {
	Name              string   `json:"name"`
	Network           string   `json:"network"`
	Asset             string   `json:"asset"`
	PayTo             string   `json:"payTo"`
	Price             string   `json:"price"`
	MaxTimeoutSeconds int      `json:"maxTimeoutSeconds,omitempty"`
	Eip712Name        string   `json:"eip712Name,omitempty"`
	Eip712Version     string   `json:"eip712Version,omitempty"`
	AddressPatterns   []string `json:"addressPatterns"`
}
