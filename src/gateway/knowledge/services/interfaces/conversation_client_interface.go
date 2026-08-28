package interfaces

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
)

type ConversationClientInterface interface {
	ListConversations(
		ctx context.Context,
		page uint,
		elementsPerPage uint,
	) (*models.ExternalListConversationsPayload, error)
	CreateConversation(
		ctx context.Context,
		body *models.ExternalCreateConversationPayload,
	) (*models.ExternalConversationPayload, error)
	UpdateConversation(
		ctx context.Context,
		conversationId string,
		body *models.ExternalUpdateConversationPayload,
	) (*models.ExternalConversationPayload, error)
	ListMessages(ctx context.Context, conversationId string) (*models.ExternalListConversationMessagesPayload, error)
	AppendMessages(
		ctx context.Context,
		conversationId string,
		body *models.ExternalAppendConversationMessagesPayload,
	) (*models.ExternalAppendConversationMessagesResultPayload, error)
}
