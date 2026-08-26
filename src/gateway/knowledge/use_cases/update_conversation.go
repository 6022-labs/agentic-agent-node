package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/responses"
)

type UpdateConversation struct {
	client interfaces.ConversationClientInterface
}

func NewUpdateConversation(client interfaces.ConversationClientInterface) *UpdateConversation {
	return &UpdateConversation{client: client}
}

func (u *UpdateConversation) Execute(
	ctx context.Context,
	request *requests.UpdateConversationRequest,
) (*responses.ConversationResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.client.UpdateConversation(
		ctx,
		request.ConversationId,
		&models.ExternalUpdateConversationPayload{Name: request.Name},
	)
	if err != nil {
		return nil, err
	}

	return &responses.ConversationResponse{
		ConversationId: payload.ConversationId,
		Name:           payload.Name,
		MessageCount:   payload.MessageCount,
		CreatedAt:      payload.CreatedAt,
		UpdatedAt:      payload.UpdatedAt,
	}, nil
}
