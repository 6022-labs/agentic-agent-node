package interfaces

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

// ErrConversationNameTaken — another conversation of the same agent already holds the name.
var ErrConversationNameTaken = errors.New("conversation name taken")

// ConversationLabelerInterface owns naming a conversation and the per-agent uniqueness it implies.
type ConversationLabelerInterface interface {
	// GetName returns the current label, or nil when the conversation is untitled or absent.
	GetName(ctx context.Context, conversationId string) (*string, error)
	// SetName fails with ErrConversationNameTaken rather than picking another name.
	SetName(ctx context.Context, conversationId string, name string) (*models.Conversation, error)
	// SetAvailableName suffixes until the name is free, so background labelling never fails on a duplicate.
	SetAvailableName(ctx context.Context, conversationId string, name string) (*models.Conversation, error)
}
