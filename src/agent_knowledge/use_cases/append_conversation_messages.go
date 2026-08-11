package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/samber/lo"
)

type AppendConversationMessages struct {
	messageAppender interfaces.MessageAppenderInterface
	agentRepository identity_repositories.AgentRepositoryInterface
}

func NewAppendConversationMessages(
	messageAppender interfaces.MessageAppenderInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *AppendConversationMessages {
	return &AppendConversationMessages{messageAppender: messageAppender, agentRepository: agentRepository}
}

func (u *AppendConversationMessages) Execute(
	ctx context.Context,
	conversationId string,
	req *requests.AppendConversationMessagesRequest,
) (*responses.AppendConversationMessagesResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("append_failed", err.Error())
	}
	if agent == nil {
		return nil, errors.NewInternalError("append_failed", "agent identity is not initialized yet")
	}
	// Seeding memory is refused until mint, like the payment policy.
	if agent.CollectionTokenId == nil {
		return nil, errors.NewConflictError("agent_not_minted", "agent identity is not minted yet")
	}

	domainMessages := make([]models.ConversationMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		domainMessages = append(domainMessages, models.ConversationMessage{
			Role:      enums.ConversationMessageRole(m.Role),
			Content:   lo.EmptyableToPtr(m.Content),
			Timestamp: m.Timestamp,
		})
	}

	inserted, err := u.messageAppender.Append(ctx, conversationId, domainMessages)
	if err != nil {
		return nil, errors.NewInternalError("append_failed", err.Error())
	}

	out := make([]responses.ConversationMessage, 0, len(inserted))
	for _, m := range inserted {
		var wire responses.ConversationMessage
		wire.FromModel(m)
		out = append(out, wire)
	}
	return &responses.AppendConversationMessagesResponse{Inserted: out}, nil
}
