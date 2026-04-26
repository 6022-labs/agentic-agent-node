package responses

type GetLlmConfigResponse struct {
	Provider   string         `json:"provider"`
	Model      string         `json:"model"`
	BaseUrl    string         `json:"baseUrl,omitempty"`
	Extensions map[string]any `json:"extensions,omitempty"`
}
