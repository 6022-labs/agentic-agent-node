package models

// ExternalA2ARequest is the JSON-RPC 2.0 envelope an A2A server expects.
type ExternalA2ARequest struct {
	JsonRpc string                       `json:"jsonrpc"`
	Id      string                       `json:"id"`
	Method  string                       `json:"method"`
	Params  ExternalA2AMessageSendParams `json:"params"`
}

type ExternalA2AMessageSendParams struct {
	Message ExternalA2AMessage `json:"message"`
}

type ExternalA2AMessage struct {
	MessageId string            `json:"messageId"`
	Role      string            `json:"role"`
	Parts     []ExternalA2APart `json:"parts"`
	ContextId string            `json:"contextId,omitempty"`
}

type ExternalA2APart struct {
	Text string `json:"text"`
}
