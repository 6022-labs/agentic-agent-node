package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
)

type MessageAppenderInterface interface {
	// Append demojizes, hashes and dedups against stored rows, returning only
	// the rows actually inserted.
	Append(
		ctx context.Context,
		conversationId string,
		messages []models.ConversationMessage,
	) ([]models.ConversationMessage, error)
}
