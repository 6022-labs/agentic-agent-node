package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

// ErrConversationNameTaken — the caller already has a conversation with that name.
var ErrConversationNameTaken = errors.New("conversation name taken")

// ConversationLabelerInterface owns conversation naming and its per-caller uniqueness.
type ConversationLabelerInterface interface {
	// GetName returns the current label, or nil when the conversation is untitled or absent.
	GetName(ctx context.Context, conversationId string) (*string, error)
	// SetName fails with ErrConversationNameTaken rather than picking another name.
	SetName(ctx context.Context, conversationId string, name string) (*models.Conversation, error)
	// SetAvailableName suffixes until the name is free, so background labelling never fails on a duplicate.
	SetAvailableName(ctx context.Context, conversationId string, name string) (*models.Conversation, error)
}
