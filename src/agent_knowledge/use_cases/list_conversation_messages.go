package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

type ListConversationMessages struct {
	conversationRepository        repositories.ConversationRepositoryInterface
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface
	agentRepository               identity_repositories.AgentRepositoryInterface
}

func NewListConversationMessages(
	conversationRepository repositories.ConversationRepositoryInterface,
	conversationMessageRepository repositories.ConversationMessageRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *ListConversationMessages {
	return &ListConversationMessages{
		conversationRepository:        conversationRepository,
		conversationMessageRepository: conversationMessageRepository,
		agentRepository:               agentRepository,
	}
}

// Execute reads a transcript; one out of the caller's view is reported absent, never forbidden.
func (u *ListConversationMessages) Execute(
	ctx context.Context,
	request *requests.ListConversationMessagesRequest,
) (*responses.ListConversationMessagesResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	callerId, err := helpers.RequireCallerId(ctx)
	if err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}
	if agent == nil {
		return nil, errors.NewInternalError("list_failed", errors.ErrAgentNotInitialized)
	}

	scope := &callerId
	if agent.IsOwnedBy(callerId) {
		scope = nil
	}
	conversation, err := u.conversationRepository.GetByIdAndAgentIdAndCallerId(
		ctx,
		request.ConversationId,
		agent.Id,
		scope,
	)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}
	if conversation == nil {
		return nil, errors.NewNotFoundError("conversation_not_found", "conversation not found")
	}

	messages, err := u.conversationMessageRepository.ListByConversationId(ctx, conversation.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err)
	}

	out := make([]responses.ConversationMessage, 0, len(messages))
	for _, m := range messages {
		var wire responses.ConversationMessage
		wire.FromModel(m)
		out = append(out, wire)
	}
	return &responses.ListConversationMessagesResponse{Messages: out}, nil
}
