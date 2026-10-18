package interfaces

import (
	"context"
	"errors"
)

// ErrConversationUnavailable — the requested key belongs to another caller.
var ErrConversationUnavailable = errors.New("conversation unavailable for caller")

// ErrAgentNotMinted — no sessions before mint, like append.
var ErrAgentNotMinted = errors.New("agent identity is not minted yet")

// ConversationResolverInterface binds a conversation key to the caller driving the turn.
type ConversationResolverInterface interface {
	// Resolve returns the stored id to use: claims a free key (empty = new session) or resumes the caller's own;
	// a key held by another caller fails with ErrConversationUnavailable, except for the agent owner.
	Resolve(ctx context.Context, requestedId string, callerId string) (string, error)
}
