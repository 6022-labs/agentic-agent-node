package use_cases

import (
	"context"
	"sort"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListConversations struct {
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
	conversationRepository        repositories.ConversationRepositoryInterface
	agentRepository               identity_repositories.AgentRepositoryInterface
}

func NewListConversations(
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface,
	conversationRepository repositories.ConversationRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *ListConversations {
	return &ListConversations{
		conversationMessageRepository: conversationMessageRepository,
		conversationRepository:        conversationRepository,
		agentRepository:               agentRepository,
	}
}

// Execute lists the agent's conversation heads, summarises each via
// ListByConversation, and orders by most recent activity.
func (u *ListConversations) Execute(ctx context.Context) (*responses.ListConversationsResponse, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}
	if agent == nil {
		return &responses.ListConversationsResponse{Conversations: []responses.ConversationSummary{}}, nil
	}

	conversations, err := u.conversationRepository.ListByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	out := make([]responses.ConversationSummary, 0, len(conversations))
	for _, conversation := range conversations {
		messages, err := u.conversationMessageRepository.ListByConversationId(ctx, conversation.Id)
		if err != nil {
			return nil, errors.NewInternalError("list_failed", err.Error())
		}
		if len(messages) == 0 {
			continue
		}
		// ListByConversation returns messages ordered by timestamp ascending.
		out = append(out, responses.ConversationSummary{
			ConversationId: conversation.Id,
			MessageCount:   int64(len(messages)),
			FirstMessageAt: messages[0].Timestamp,
			LastMessageAt:  messages[len(messages)-1].Timestamp,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].LastMessageAt.After(out[j].LastMessageAt)
	})
	return &responses.ListConversationsResponse{Conversations: out}, nil
}
