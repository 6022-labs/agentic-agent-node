package models

type ExternalAnthropicMessagesStandardResponse struct {
	Model string                                         `json:"model"`
	ID    string                                         `json:"id"`
	Usage ExternalAnthropicMessagesStandardResponseUsage `json:"usage"`
	ExternalAnthropicMessagesStandardMessage
}

type ExternalAnthropicMessagesStandardResponseUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
