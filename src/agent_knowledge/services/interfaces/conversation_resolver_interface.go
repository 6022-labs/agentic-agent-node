package interfaces

import (
	"context"
)

// ConversationResolverInterface binds a conversation key to the caller driving the turn.
type ConversationResolverInterface interface {
	// Resolve returns the stored id to use: claims a free key (empty = new session) or resumes the caller's own;
	// a key held by another caller fails with ErrConversationUnavailable, except for the agent owner.
	Resolve(ctx context.Context, requestedId string, callerId string) (string, error)
}
