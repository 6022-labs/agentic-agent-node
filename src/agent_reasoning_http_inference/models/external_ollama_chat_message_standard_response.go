package models

type ExternalOllamaChatMessageStandardResponse struct {
	Model      string                                   `json:"model"`
	Message    ExternalOllamaChatMessageStandardMessage `json:"message"`
	DoneReason *string                                  `json:"done_reason,omitempty"`
	// Ollama reports token counts at the top level of the final response object.
	PromptEvalCount int `json:"prompt_eval_count"`
	EvalCount       int `json:"eval_count"`
}
