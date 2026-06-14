package responses

import "time"

// GetLlmConfigResponse is the gateway-shaped wire response for LLM config
// endpoints. The api key is never exposed.
type GetLlmConfigResponse struct {
	Id          string    `json:"id"`
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
	BaseUrl     string    `json:"baseUrl,omitempty"`
	SetupStatus string    `json:"setupStatus"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
