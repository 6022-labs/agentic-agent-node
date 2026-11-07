package services

import (
	"context"
	"strings"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// ConversationResolver keeps sessions private to their caller; the agent owner resumes any.
type ConversationResolver struct {
	conversationRepository repositories.ConversationRepositoryInterface
	agentRepository        identity_repositories.AgentRepositoryInterface
	knowledgeMetrics       metrics_interfaces.KnowledgeMetricsInterface
}

func NewConversationResolver(
	conversationRepository repositories.ConversationRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
	knowledgeMetrics metrics_interfaces.KnowledgeMetricsInterface,
) *ConversationResolver {
	return &ConversationResolver{
		conversationRepository: conversationRepository,
		agentRepository:        agentRepository,
		knowledgeMetrics:       knowledgeMetrics,
	}
}

func (r *ConversationResolver) Resolve(ctx context.Context, requestedId string, callerId string) (string, error) {
	agent, err := r.agentRepository.Get(ctx)
	if err != nil {
		return "", err
	}
	if agent == nil || agent.CollectionTokenId == nil {
		return "", apperrors.ErrAgentNotMinted
	}

	requestedId = strings.TrimSpace(requestedId)
	if requestedId == "" {
		requestedId = uuid.NewString()
	}

	usable, err := r.claim(ctx, agent, requestedId, callerId)
	if err != nil {
		return "", err
	}
	if !usable {
		return "", apperrors.ErrConversationUnavailable
	}
	return requestedId, nil
}

// claim creates the head or reports whether the existing one is resumable by the caller.
func (r *ConversationResolver) claim(
	ctx context.Context,
	agent *identity_models.Agent,
	conversationId string,
	callerId string,
) (bool, error) {
	created, err := r.conversationRepository.Ensure(ctx, &models.Conversation{
		Id:       conversationId,
		AgentId:  agent.Id,
		CallerId: &callerId,
	})
	if err != nil {
		return false, err
	}
	if created {
		r.knowledgeMetrics.RecordConversationStarted(ctx)
		return true, nil
	}
	return r.isResumableBy(ctx, agent, conversationId, callerId)
}

// isResumableBy is false for caller-less heads unless the caller owns the agent.
func (r *ConversationResolver) isResumableBy(
	ctx context.Context,
	agent *identity_models.Agent,
	conversationId string,
	callerId string,
) (bool, error) {
	conversation, err := r.conversationRepository.GetById(ctx, conversationId)
	if err != nil {
		return false, err
	}
	if conversation == nil || conversation.AgentId != agent.Id {
		return false, nil
	}
	if agent.IsOwnedBy(callerId) {
		return true, nil
	}
	return conversation.CallerId != nil && *conversation.CallerId == callerId, nil
}
