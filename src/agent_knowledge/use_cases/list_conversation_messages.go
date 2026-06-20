package use_cases

import (
	"context"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListConversationMessages struct {
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
}

func NewListConversationMessages(
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface,
) *ListConversationMessages {
	return &ListConversationMessages{
		conversationMessageRepository: conversationMessageRepository,
	}
}

func (u *ListConversationMessages) Execute(ctx context.Context, conversationId string) (*responses.ListConversationMessagesResponse, error) {
	conversationId = strings.TrimSpace(conversationId)
	if len(conversationId) == 0 {
		return &responses.ListConversationMessagesResponse{Messages: toWireMessages(nil)}, nil
	}

	messages, err := u.conversationMessageRepository.ListByConversationId(ctx, conversationId)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}
	return &responses.ListConversationMessagesResponse{Messages: toWireMessages(messages)}, nil
}
