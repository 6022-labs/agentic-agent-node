package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/responses"
)

type ListConversationMessages struct {
	client interfaces.ConversationClientInterface
}

func NewListConversationMessages(client interfaces.ConversationClientInterface) *ListConversationMessages {
	return &ListConversationMessages{client: client}
}

func (u *ListConversationMessages) Execute(
	ctx context.Context,
	request *requests.ListConversationMessagesRequest,
) (*responses.ListConversationMessagesResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.client.ListMessages(ctx, request.ConversationId)
	if err != nil {
		return nil, err
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
