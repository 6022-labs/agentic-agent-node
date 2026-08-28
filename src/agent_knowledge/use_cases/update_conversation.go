package use_cases

import (
	"context"
	stderrors "errors"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type UpdateConversation struct {
	conversationRepository repositories.ConversationRepositoryInterface
	conversationLabeler    interfaces.ConversationLabelerInterface
	agentRepository        identity_repositories.AgentRepositoryInterface
}

func NewUpdateConversation(
	conversationRepository repositories.ConversationRepositoryInterface,
	conversationLabeler interfaces.ConversationLabelerInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *UpdateConversation {
	return &UpdateConversation{
		conversationRepository: conversationRepository,
		conversationLabeler:    conversationLabeler,
		agentRepository:        agentRepository,
	}
}

// Execute relabels a memory session; another agent's conversation is reported as absent, never as forbidden.
func (u *UpdateConversation) Execute(
	ctx context.Context,
	request *requests.UpdateConversationRequest,
) (*responses.ConversationResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if agent == nil {
		return nil, errors.NewInternalError("update_failed", "agent identity is not initialized yet")
	}

	conversation, err := u.conversationRepository.GetById(ctx, request.ConversationId)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if conversation == nil || conversation.AgentId != agent.Id {
		return nil, errors.NewNotFoundError("conversation_not_found", "conversation not found")
	}

	updated, err := u.conversationLabeler.SetName(ctx, conversation.Id, request.Name)
	if stderrors.Is(err, interfaces.ErrConversationNameTaken) {
		return nil, errors.NewConflictError("conversation_name_taken", "a conversation already uses that name")
	}
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}

	var wire responses.ConversationResponse
	wire.FromModel(*updated)
	return &wire, nil
}
