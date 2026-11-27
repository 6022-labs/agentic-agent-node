package models

import "strings"

// ExternalA2AResponse is the JSON-RPC reply to SendMessage; Error is set instead of Result on a protocol failure.
type ExternalA2AResponse struct {
	JsonRpc string                     `json:"jsonrpc"`
	Id      any                        `json:"id"`
	Result  *ExternalSendMessageResult `json:"result,omitempty"`
	Error   *ExternalA2AError          `json:"error,omitempty"`
}

type ExternalSendMessageResult struct {
	Message *ExternalA2AMessage `json:"message,omitempty"`
}

type ExternalA2AError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Text joins the text of every part; a reply carrying no message is empty.
func (r *ExternalA2AResponse) Text() string {
	if r.Result == nil || r.Result.Message == nil {
		return ""
	}
	var builder strings.Builder
	for _, part := range r.Result.Message.Parts {
		builder.WriteString(part.Text)
	}
	return builder.String()
}

// ContextId is the conversation the peer threaded the reply onto, if any.
func (r *ExternalA2AResponse) ContextId() string {
	if r.Result == nil || r.Result.Message == nil {
		return ""
	}
	return r.Result.Message.ContextId
}
