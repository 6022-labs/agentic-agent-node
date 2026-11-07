package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/common/helpers"
	"github.com/samber/lo"
)

type AppendConversationMessages struct {
	messageAppender        interfaces.MessageAppenderInterface
	conversationRepository repositories.ConversationRepositoryInterface
	agentRepository        identity_repositories.AgentRepositoryInterface
	knowledgeMetrics       metrics_interfaces.KnowledgeMetricsInterface
}

func NewAppendConversationMessages(
	messageAppender interfaces.MessageAppenderInterface,
	conversationRepository repositories.ConversationRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
	knowledgeMetrics metrics_interfaces.KnowledgeMetricsInterface,
) *AppendConversationMessages {
	return &AppendConversationMessages{
		messageAppender:        messageAppender,
		conversationRepository: conversationRepository,
		agentRepository:        agentRepository,
		knowledgeMetrics:       knowledgeMetrics,
	}
}

func (u *AppendConversationMessages) Execute(
	ctx context.Context,
	request *requests.AppendConversationMessagesRequest,
) (*responses.AppendConversationMessagesResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}
	callerId, err := helpers.RequireCallerId(ctx)
	if err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("append_failed", err)
	}
	if agent == nil {
		return nil, errors.NewInternalError("append_failed", errors.ErrAgentNotInitialized)
	}
	// Seeding memory is refused until mint, like the payment policy.
	if agent.CollectionTokenId == nil {
		return nil, errors.NewConflictError("agent_not_minted", "agent identity is not minted yet")
	}

	// An existing head keeps its caller.
	created, err := u.conversationRepository.Ensure(ctx, &models.Conversation{
		Id:       request.ConversationId,
		AgentId:  agent.Id,
		CallerId: &callerId,
	})
	if err != nil {
		return nil, errors.NewInternalError("append_failed", err)
	}
	if created {
		u.knowledgeMetrics.RecordConversationStarted(ctx)
	}

	// The owner may seed any session; anyone else only their own.
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
		return nil, errors.NewInternalError("append_failed", err)
	}
	if conversation == nil {
		return nil, errors.NewNotFoundError("conversation_not_found", "conversation not found")
	}

	domainMessages := make([]models.ConversationMessage, 0, len(request.Messages))
	for _, m := range request.Messages {
		domainMessages = append(domainMessages, models.ConversationMessage{
			Role:      enums.ConversationMessageRole(m.Role),
			Content:   lo.EmptyableToPtr(m.Content),
			Timestamp: m.Timestamp,
		})
	}

	inserted, err := u.messageAppender.Append(ctx, request.ConversationId, domainMessages)
	if err != nil {
		return nil, errors.NewInternalError("append_failed", err)
	}

	out := make([]responses.ConversationMessage, 0, len(inserted))
	for _, m := range inserted {
		var wire responses.ConversationMessage
		wire.FromModel(m)
		out = append(out, wire)
	}
	return &responses.AppendConversationMessagesResponse{Inserted: out}, nil
}
