package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/google/uuid"
)

type CreateConversation struct {
	conversationRepository repositories.ConversationRepositoryInterface
	agentRepository        identity_repositories.AgentRepositoryInterface
	knowledgeMetrics       metrics_interfaces.KnowledgeMetricsInterface
}

func NewCreateConversation(
	conversationRepository repositories.ConversationRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
	knowledgeMetrics metrics_interfaces.KnowledgeMetricsInterface,
) *CreateConversation {
	return &CreateConversation{
		conversationRepository: conversationRepository,
		agentRepository:        agentRepository,
		knowledgeMetrics:       knowledgeMetrics,
	}
}

// Execute opens an empty named session; the id is server-generated.
func (u *CreateConversation) Execute(
	ctx context.Context,
	request *requests.CreateConversationRequest,
) (*responses.ConversationResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	callerId, err := helpers.RequireCallerId(ctx)
	if err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	if agent == nil {
		return nil, errors.NewInternalError("create_failed", errors.ErrAgentNotInitialized)
	}
	if agent.CollectionTokenId == nil {
		return nil, errors.NewConflictError("agent_not_minted", "agent identity is not minted yet")
	}

	taken, err := u.conversationRepository.ExistsByAgentIdAndCallerIdAndName(
		ctx,
		agent.Id,
		&callerId,
		request.Name,
		nil,
	)
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}
	if taken {
		return nil, errors.NewConflictError("conversation_name_taken", "a conversation already uses that name")
	}

	created, err := u.conversationRepository.Create(ctx, &models.Conversation{
		Id:       uuid.NewString(),
		AgentId:  agent.Id,
		CallerId: &callerId,
		Name:     &request.Name,
	})
	if err != nil {
		return nil, errors.NewInternalError("create_failed", err)
	}

	u.knowledgeMetrics.RecordConversationStarted(ctx)
	u.knowledgeMetrics.RecordConversationNamed(ctx)

	var wire responses.ConversationResponse
	wire.FromModel(*created)
	return &wire, nil
}
