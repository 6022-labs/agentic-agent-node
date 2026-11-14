package use_cases

import (
	"context"
	stderrors "errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
)

type UpdateConversation struct {
	conversationRepository repositories.ConversationRepositoryInterface
	conversationLabeler    interfaces.ConversationLabelerInterface
	agentResolver          identity_interfaces.AgentResolverInterface
}

func NewUpdateConversation(
	conversationRepository repositories.ConversationRepositoryInterface,
	conversationLabeler interfaces.ConversationLabelerInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *UpdateConversation {
	return &UpdateConversation{
		conversationRepository: conversationRepository,
		conversationLabeler:    conversationLabeler,
		agentResolver:          agentResolver,
	}
}

// Execute relabels a session; one out of the caller's view is reported absent, never forbidden.
func (u *UpdateConversation) Execute(
	ctx context.Context,
	request *requests.UpdateConversationRequest,
) (*responses.ConversationResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	callerId, err := helpers.RequireCallerId(ctx)
	if err != nil {
		return nil, err
	}

	agent, err := u.agentResolver.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err)
	}
	if agent == nil {
		return nil, errors.NewInternalError("update_failed", errors.ErrAgentNotInitialized)
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
		return nil, errors.NewInternalError("update_failed", err)
	}
	if conversation == nil {
		return nil, errors.NewNotFoundError("conversation_not_found", "conversation not found")
	}

	updated, err := u.conversationLabeler.SetName(ctx, conversation.Id, request.Name)
	if stderrors.Is(err, interfaces.ErrConversationNameTaken) {
		return nil, errors.NewConflictError("conversation_name_taken", "a conversation already uses that name")
	}
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err)
	}

	var wire responses.ConversationResponse
	wire.FromModel(*updated)
	return &wire, nil
}
