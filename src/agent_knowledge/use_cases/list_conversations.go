package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/samber/lo"
)

type ListConversations struct {
	conversationRepository        repositories.ConversationRepositoryInterface
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
	agentResolver                 identity_interfaces.AgentResolverInterface
}

func NewListConversations(
	conversationRepository repositories.ConversationRepositoryInterface,
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *ListConversations {
	return &ListConversations{
		conversationRepository:        conversationRepository,
		conversationMessageRepository: conversationMessageRepository,
		agentResolver:                 agentResolver,
	}
}

// Execute pages sessions newest first with message counts; the agent owner sees every caller's.
func (u *ListConversations) Execute(
	ctx context.Context,
	request *requests.ListConversationsRequest,
) (*responses.ListConversationsResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	callerId, err := helpers.RequireCallerId(ctx)
	if err != nil {
		return nil, err
	}

	agent, err := u.agentResolver.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}
	if agent == nil {
		return &responses.ListConversationsResponse{
			Page:            *request.Page,
			ElementsPerPage: *request.ElementsPerPage,
			Conversations:   []responses.ConversationResponse{},
		}, nil
	}

	scope := &callerId
	if agent.IsOwnedBy(callerId) {
		scope = nil
	}

	total, err := u.conversationRepository.CountByAgentId(ctx, agent.Id, scope)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}

	limit := int(*request.ElementsPerPage)
	offset := int(*request.Page-1) * limit

	conversations, err := u.conversationRepository.ListByAgentId(ctx, agent.Id, scope, &offset, &limit)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}

	counts, err := u.conversationMessageRepository.CountByConversationIds(ctx, lo.Map(
		conversations,
		func(conversation models.Conversation, _ int) string { return conversation.Id },
	))
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
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
