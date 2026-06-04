package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListConversationMessagesInterface interface {
	Execute(ctx context.Context, conversationId string) (*responses.ListConversationMessagesResponse, error)
}

type ListConversationMessages struct {
	service interfaces.ListConversationMessagesInterface
}

func NewListConversationMessages(service interfaces.ListConversationMessagesInterface) *ListConversationMessages {
	return &ListConversationMessages{service: service}
}

func (u *ListConversationMessages) Execute(ctx context.Context, conversationId string) (*responses.ListConversationMessagesResponse, error) {
	messages, err := u.service.List(ctx, conversationId)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}
	return &responses.ListConversationMessagesResponse{Messages: toWireMessages(messages)}, nil
}
