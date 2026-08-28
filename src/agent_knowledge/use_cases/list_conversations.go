package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/samber/lo"
)

type ListConversations struct {
	conversationRepository        repositories.ConversationRepositoryInterface
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
	agentRepository               identity_repositories.AgentRepositoryInterface
}

func NewListConversations(
	conversationRepository repositories.ConversationRepositoryInterface,
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *ListConversations {
	return &ListConversations{
		conversationRepository:        conversationRepository,
		conversationMessageRepository: conversationMessageRepository,
		agentRepository:               agentRepository,
	}
}

// Execute returns one page of the agent's memory sessions, newest first, decorated with their message counts.
func (u *ListConversations) Execute(
	ctx context.Context,
	request *requests.ListConversationsRequest,
) (*responses.ListConversationsResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}
	if agent == nil {
		return &responses.ListConversationsResponse{
			Page:            *request.Page,
			ElementsPerPage: *request.ElementsPerPage,
			Conversations:   []responses.ConversationResponse{},
		}, nil
	}

	total, err := u.conversationRepository.CountByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	limit := int(*request.ElementsPerPage)
	offset := int(*request.Page-1) * limit

	conversations, err := u.conversationRepository.ListByAgentId(ctx, agent.Id, &offset, &limit)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	counts, err := u.conversationMessageRepository.CountByConversationIds(ctx, lo.Map(
		conversations,
		func(conversation models.Conversation, _ int) string { return conversation.Id },
	))
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}

	out := make([]responses.ConversationResponse, 0, len(conversations))
	for _, conversation := range conversations {
		var wire responses.ConversationResponse
		wire.FromModel(conversation)
		wire.MessageCount = counts[conversation.Id]
		out = append(out, wire)
	}

	return &responses.ListConversationsResponse{
		Page:            *request.Page,
		ElementsPerPage: *request.ElementsPerPage,
		TotalElements:   total,
		Conversations:   out,
	}, nil
}
