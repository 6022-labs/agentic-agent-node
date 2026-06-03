package interfaces

import (
	"context"
	"time"
)

type ConversationHistoryInterface interface {
	// Read returns every message for a conversation ordered by timestamp asc.
	Read(ctx context.Context, conversationId string) ([]ConversationMessage, error)
	// Write persists messages and returns the rows actually inserted.
	Write(ctx context.Context, conversationId string, messages []ConversationMessage) ([]ConversationMessage, error)
}

// ConversationMessage is the port-level DTO — semantic fields only, no
// persistence concerns (id, hash, created-at).
type ConversationMessage struct {
	Role      string
	Content   string
	Timestamp time.Time
}
