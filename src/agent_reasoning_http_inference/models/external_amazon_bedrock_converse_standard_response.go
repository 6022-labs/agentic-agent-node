package models

type ExternalAmazonBedrockConverseStandardResponse struct {
	Output     ExternalAmazonBedrockConverseStandardResponseOutput `json:"output"`
	StopReason *string                                             `json:"stopReason,omitempty"`
	Usage      ExternalAmazonBedrockConverseStandardResponseUsage  `json:"usage"`
}

type ExternalAmazonBedrockConverseStandardResponseUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
	TotalTokens  int `json:"totalTokens"`
}

type ExternalAmazonBedrockConverseStandardResponseOutput struct {
	Message ExternalAmazonBedrockConverseStandardMessage `json:"message"`
}
