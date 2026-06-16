package models

// The /internal/responses contract mirrors the OpenAI Responses shape so the
// gateway and runtime speak one dialect; the gateway only authenticates and
// forwards.

type ExternalInputMessagePayload struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ExternalConversationRefPayload struct {
	Id string `json:"id"`
}

type ExternalCreateResponsePayload struct {
	Model              string                          `json:"model,omitempty"`
	Instructions       string                          `json:"instructions,omitempty"`
	Input              []ExternalInputMessagePayload   `json:"input"`
	Conversation       *ExternalConversationRefPayload `json:"conversation,omitempty"`
	PreviousResponseId string                          `json:"previous_response_id,omitempty"`
	Temperature        *float64                        `json:"temperature,omitempty"`
	TopP               *float64                        `json:"top_p,omitempty"`
	MaxOutputTokens    *int                            `json:"max_output_tokens,omitempty"`
	Store              *bool                           `json:"store,omitempty"`
	Stream             bool                            `json:"stream,omitempty"`
	Metadata           map[string]string               `json:"metadata,omitempty"`
}

type ExternalOutputContentPayload struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []any  `json:"annotations"`
}

type ExternalOutputItemPayload struct {
	Type    string                         `json:"type"`
	Id      string                         `json:"id"`
	Status  string                         `json:"status"`
	Role    string                         `json:"role"`
	Content []ExternalOutputContentPayload `json:"content"`
}

type ExternalUsagePayload struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type ExternalResponseErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ExternalCreateResponseResultPayload struct {
	Id                 string                          `json:"id"`
	Object             string                          `json:"object"`
	CreatedAt          int64                           `json:"created_at"`
	Status             string                          `json:"status"`
	Model              string                          `json:"model,omitempty"`
	Output             []ExternalOutputItemPayload     `json:"output"`
	Usage              *ExternalUsagePayload           `json:"usage"`
	PreviousResponseId *string                         `json:"previous_response_id"`
	Conversation       *ExternalConversationRefPayload `json:"conversation"`
	Metadata           map[string]string               `json:"metadata,omitempty"`
	Error              *ExternalResponseErrorPayload   `json:"error"`
}
