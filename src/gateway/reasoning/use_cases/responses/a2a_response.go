package responses

import "encoding/json"

// A2AResponse is the JSON-RPC 2.0 success envelope; failures travel as HTTP error status instead.
type A2AResponse struct {
	JsonRpc string             `json:"jsonrpc"`
	Id      json.RawMessage    `json:"id,omitempty"`
	Result  *SendMessageResult `json:"result,omitempty"`
}

// SendMessageResult is the ProtoJSON of SendMessageResponse's oneof; we always return the message variant.
type SendMessageResult struct {
	Message *A2AMessage `json:"message,omitempty"`
}

type A2AMessage struct {
	MessageId string    `json:"messageId"`
	Role      string    `json:"role"`
	Parts     []A2APart `json:"parts"`
	ContextId string    `json:"contextId,omitempty"`
}

type A2APart struct {
	Text string `json:"text"`
}
