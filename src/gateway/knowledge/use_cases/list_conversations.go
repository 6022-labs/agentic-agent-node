package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/knowledge/use_cases/responses"
)

type ListConversations struct {
	client interfaces.ConversationClientInterface
}

func NewListConversations(client interfaces.ConversationClientInterface) *ListConversations {
	return &ListConversations{client: client}
}

func (u *ListConversations) Execute(
	ctx context.Context,
	request *requests.ListConversationsRequest,
) (*responses.ListConversationsResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.client.ListConversations(ctx, *request.Page, *request.ElementsPerPage)
	if err != nil {
		return nil, errors.NewUnavailableError("runtime_unavailable", err.Error())
	}

	out := make([]responses.ConversationResponse, 0, len(payload.Conversations))
	for _, c := range payload.Conversations {
		out = append(out, responses.ConversationResponse{
			ConversationId: c.ConversationId,
			Name:           c.Name,
			MessageCount:   c.MessageCount,
			CreatedAt:      c.CreatedAt,
			UpdatedAt:      c.UpdatedAt,
		})
	}
	return &responses.ListConversationsResponse{
		Page:            payload.Page,
		ElementsPerPage: payload.ElementsPerPage,
		TotalElements:   payload.TotalElements,
		Conversations:   out,
	}, nil
}
