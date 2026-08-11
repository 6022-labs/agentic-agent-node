package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/responses"
)

type ListConversationMessages struct {
	client interfaces.ConversationHistoryClientInterface
}

func NewListConversationMessages(client interfaces.ConversationHistoryClientInterface) *ListConversationMessages {
	return &ListConversationMessages{client: client}
}

func (u *ListConversationMessages) Execute(
	ctx context.Context,
	conversationId string,
) (*responses.ListConversationMessagesResponse, error) {
	payload, err := u.client.ListMessages(ctx, conversationId)
	if err != nil {
		return nil, errors.NewUnavailableError("runtime_unavailable", err.Error())
	}

	out := make([]responses.ConversationMessage, 0, len(payload.Messages))
	for _, m := range payload.Messages {
		out = append(out, responses.ConversationMessage{
			Id:             m.Id,
			ConversationId: m.ConversationId,
			Role:           m.Role,
			Content:        m.Content,
			ToolCalls:      m.ToolCalls,
			ToolCallId:     m.ToolCallId,
			Timestamp:      m.Timestamp,
		})
	}
	return &responses.ListConversationMessagesResponse{Messages: out}, nil
}
