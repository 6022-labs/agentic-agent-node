package models

import "time"

// LLMConnectionPayload mirrors the JSON shape the runtime returns for
// /internal/config/llm. Kept independent of the runtime's domain types.
type LLMConnectionPayload struct {
	Id          string    `json:"id"`
	Platform    string    `json:"platform"`
	Model       string    `json:"model"`
	BaseUrl     *string   `json:"baseUrl,omitempty"`
	SetupStatus string    `json:"setupStatus"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// SetupLLMConnectionPayload is the body shape /internal/config/llm POST accepts.
type SetupLLMConnectionPayload struct {
	Platform string  `json:"platform"`
	Model    string  `json:"model"`
	BaseUrl  *string `json:"baseUrl,omitempty"`
	ApiKey   string  `json:"apiKey"`
}
