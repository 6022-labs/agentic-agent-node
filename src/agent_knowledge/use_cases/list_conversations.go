package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListConversations struct {
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
}

func NewListConversations(conversationMessageRepository repositories.ConversationMessageRepositoryInterface) *ListConversations {
	return &ListConversations{conversationMessageRepository: conversationMessageRepository}
}

// Execute lists the ids, then summarises each via ListByConversation (one call per conversation).
func (u *ListConversations) Execute(ctx context.Context) (*responses.ListConversationsResponse, error) {
	ids, err := u.conversationMessageRepository.ListConversationIds(ctx)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	out := make([]responses.ConversationSummary, 0, len(ids))
	for _, id := range ids {
		messages, err := u.conversationMessageRepository.ListByConversation(ctx, id)
		if err != nil {
			return nil, errors.NewInternalError("list_failed", err.Error())
		}
		if len(messages) == 0 {
			continue
		}
		// ListByConversation returns messages ordered by timestamp ascending.
		out = append(out, responses.ConversationSummary{
			ConversationId: id,
			MessageCount:   int64(len(messages)),
			FirstMessageAt: messages[0].Timestamp,
			LastMessageAt:  messages[len(messages)-1].Timestamp,
		})
	}
	return &responses.ListConversationsResponse{Conversations: out}, nil
}
