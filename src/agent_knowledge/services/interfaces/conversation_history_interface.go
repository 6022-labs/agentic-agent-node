package interfaces

import "time"

type ConversationHistoryInterface interface {
	// Read returns every message for a conversation ordered by timestamp asc.
	Read(conversationId string) ([]ConversationMessage, error)
	// Write persists messages and returns the rows actually inserted.
	Write(conversationId string, messages []ConversationMessage) ([]ConversationMessage, error)
}

// ConversationMessage is the port-level DTO — semantic fields only, no
// persistence concerns (id, hash, created-at).
type ConversationMessage struct {
	Role      string
	Author    string
	Platform  string
	Content   string
	Timestamp time.Time
}
