package use_cases

import (
	"context"

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
}

func NewAppendConversationMessages(messageAppender interfaces.MessageAppenderInterface) *AppendConversationMessages {
	return &AppendConversationMessages{messageAppender: messageAppender}
}

func (u *AppendConversationMessages) Execute(ctx context.Context, conversationId string, req *requests.AppendConversationMessagesRequest) (*responses.AppendConversationMessagesResponse, error) {
	if err := req.ValidateAndSanitize(); err != nil {
		return nil, err
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
