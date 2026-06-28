package requests

import (
	"encoding/json"
	"strings"
)

// A2ARequest is the JSON-RPC 2.0 envelope for an A2A call; only SendMessage is handled.
type A2ARequest struct {
	JsonRpc string               `json:"jsonrpc"`
	Id      json.RawMessage      `json:"id"`
	Method  A2AMethod            `json:"method"`
	Params  A2AMessageSendParams `json:"params"`
}

type A2AMessageSendParams struct {
	Message A2AMessage `json:"message"`
}

type A2AMessage struct {
	MessageId string    `json:"messageId"`
	Role      string    `json:"role"`
	Parts     []A2APart `json:"parts"`
	ContextId string    `json:"contextId,omitempty"`
}

// A2APart is the ProtoJSON oneof; only the text variant is read (file/data variants carry no text).
type A2APart struct {
	Text string `json:"text"`
}

// Text joins the text of every part; non-text parts contribute nothing.
func (m *A2AMessage) Text() string {
	var builder strings.Builder
	for _, part := range m.Parts {
		builder.WriteString(part.Text)
	}
	return builder.String()
}
