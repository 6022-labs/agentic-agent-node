package responses

// OutputContent is one content part of an output message (OpenAI: output_text).
type OutputContent struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []any  `json:"annotations"`
}

// OutputItem is one item of the response output (today only assistant messages).
type OutputItem struct {
	Type    string          `json:"type"`
	Id      string          `json:"id"`
	Status  string          `json:"status"`
	Role    string          `json:"role"`
	Content []OutputContent `json:"content"`
}

type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type ResponseError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ConversationRef is the {id} shape OpenAI returns for `conversation`.
type ConversationRef struct {
	Id string `json:"id"`
}

// CreateResponseResponse mirrors the OpenAI Responses object. Id doubles as the
// conversation key, so callers continue a thread by passing it back as
// previous_response_id.
type CreateResponseResponse struct {
	Id                 string            `json:"id"`
	Object             string            `json:"object"`
	CreatedAt          int64             `json:"created_at"`
	Status             string            `json:"status"`
	Model              string            `json:"model,omitempty"`
	Output             []OutputItem      `json:"output"`
	Usage              *Usage            `json:"usage"`
	PreviousResponseId *string           `json:"previous_response_id"`
	Conversation       *ConversationRef  `json:"conversation"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Error              *ResponseError    `json:"error"`
}
