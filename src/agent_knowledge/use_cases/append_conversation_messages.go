package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type AppendConversationMessages struct {
	messageAppender interfaces.MessageAppenderInterface
}

func NewAppendConversationMessages(messageAppender interfaces.MessageAppenderInterface) *AppendConversationMessages {
	return &AppendConversationMessages{messageAppender: messageAppender}
}

func (u *AppendConversationMessages) Execute(ctx context.Context, conversationId string, req *requests.AppendConversationMessagesRequest) (*responses.AppendConversationMessagesResponse, error) {
	domainMessages := make([]models.ConversationMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		domainMessages = append(domainMessages, models.ConversationMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}

	inserted, err := u.messageAppender.Append(ctx, conversationId, domainMessages)
	if err != nil {
		return nil, errors.NewInternalError("append_failed", err.Error())
	}

	return &responses.AppendConversationMessagesResponse{Inserted: toWireMessages(inserted)}, nil
}

func toWireMessages(in []models.ConversationMessage) []responses.ConversationMessage {
	out := make([]responses.ConversationMessage, 0, len(in))
	for _, m := range in {
		out = append(out, responses.ConversationMessage{
			Id:             m.Id,
			ConversationId: m.ConversationId,
			Role:           m.Role,
			Content:        m.Content,
			Timestamp:      m.Timestamp,
		})
	}
	return out
}
