package models

type ExternalJsonRPCResponse[T any] struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  T             `json:"result"`
	Error   *ExternalJsonRPCError `json:"error,omitempty"`
}
