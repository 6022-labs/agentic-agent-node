package clients

import "fmt"

// LlmStatusError carries a non-2xx provider response so callers read the status, not a message.
type LlmStatusError struct {
	StatusCode int
}

func (e *LlmStatusError) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}
