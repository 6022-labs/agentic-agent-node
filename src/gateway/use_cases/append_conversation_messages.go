package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

type AppendConversationMessages struct {
	client interfaces.ConversationHistoryClientInterface
}

func NewAppendConversationMessages(client interfaces.ConversationHistoryClientInterface) *AppendConversationMessages {
	return &AppendConversationMessages{client: client}
}

func (u *AppendConversationMessages) Execute(ctx context.Context, conversationId string, request *requests.AppendConversationMessagesRequest) (*responses.AppendConversationMessagesResponse, error) {
	payload := &models.ExternalAppendConversationMessagesPayload{
		Messages: make([]models.ExternalConversationMessageInput, 0, len(request.Messages)),
	}
	for _, message := range request.Messages {
		payload.Messages = append(payload.Messages, models.ExternalConversationMessageInput{
			Role:      string(message.Role),
			Content:   message.Content,
			Timestamp: message.Timestamp,
		})
	}

	result, err := u.client.AppendMessages(ctx, conversationId, payload)
	if err != nil {
		return nil, errors.NewUnavailableError("runtime_unavailable", err.Error())
	}

	inserted := make([]responses.ConversationMessage, 0, len(result.Inserted))
	for _, message := range result.Inserted {
		inserted = append(inserted, responses.ConversationMessage{
			Id:             message.Id,
			ConversationId: message.ConversationId,
			Role:           message.Role,
			Content:        message.Content,
			Timestamp:      message.Timestamp,
		})
	}
	return &responses.AppendConversationMessagesResponse{Inserted: inserted}, nil
}
