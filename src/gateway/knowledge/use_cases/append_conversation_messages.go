package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/responses"
)

type AppendConversationMessages struct {
	client interfaces.ConversationClientInterface
}

func NewAppendConversationMessages(client interfaces.ConversationClientInterface) *AppendConversationMessages {
	return &AppendConversationMessages{client: client}
}

func (u *AppendConversationMessages) Execute(
	ctx context.Context,
	conversationId string,
	request *requests.AppendConversationMessagesRequest,
) (*responses.AppendConversationMessagesResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

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
		return nil, err
	}

	inserted := make([]responses.ConversationMessage, 0, len(result.Inserted))
	for _, message := range result.Inserted {
		inserted = append(inserted, responses.ConversationMessage{
			Id:             message.Id,
			ConversationId: message.ConversationId,
			Role:           message.Role,
			Content:        message.Content,
			ToolCalls:      message.ToolCalls,
			ToolCallId:     message.ToolCallId,
			Timestamp:      message.Timestamp,
		})
	}
	return &responses.AppendConversationMessagesResponse{Inserted: inserted}, nil
}
