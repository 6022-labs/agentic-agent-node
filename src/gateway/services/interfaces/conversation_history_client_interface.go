package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
)

type ConversationHistoryClientInterface interface {
	ListConversations(ctx context.Context) (*models.ExternalListConversationsPayload, error)
	ListMessages(ctx context.Context, conversationId string) (*models.ExternalListConversationMessagesPayload, error)
	AppendMessages(ctx context.Context, conversationId string, body *models.ExternalAppendConversationMessagesPayload) (*models.ExternalAppendConversationMessagesResultPayload, error)
}
