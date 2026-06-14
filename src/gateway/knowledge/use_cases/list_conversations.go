package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/responses"
)

type ListConversations struct {
	client interfaces.ConversationHistoryClientInterface
}

func NewListConversations(client interfaces.ConversationHistoryClientInterface) *ListConversations {
	return &ListConversations{client: client}
}

func (u *ListConversations) Execute(ctx context.Context) (*responses.ListConversationsResponse, error) {
	payload, err := u.client.ListConversations(ctx)
	if err != nil {
		return nil, errors.NewUnavailableError("runtime_unavailable", err.Error())
	}

	out := make([]responses.ConversationSummary, 0, len(payload.Conversations))
	for _, c := range payload.Conversations {
		out = append(out, responses.ConversationSummary{
			ConversationId: c.ConversationId,
			MessageCount:   c.MessageCount,
			FirstMessageAt: c.FirstMessageAt,
			LastMessageAt:  c.LastMessageAt,
		})
	}
	return &responses.ListConversationsResponse{Conversations: out}, nil
}
