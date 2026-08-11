package use_cases

import (
	"context"
	"errors"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type UpdateSelfTool struct {
	selfToolConfigClient gateway_interfaces.SelfToolConfigClientInterface
}

func NewUpdateSelfTool(selfToolConfigClient gateway_interfaces.SelfToolConfigClientInterface) *UpdateSelfTool {
	return &UpdateSelfTool{selfToolConfigClient: selfToolConfigClient}
}

func (u *UpdateSelfTool) Execute(ctx context.Context, request *requests.UpdateSelfToolRequest) (*responses.SelfToolResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.selfToolConfigClient.Update(ctx, request.Id, &models.ExternalUpdateSelfToolPayload{
		Enabled: *request.Enabled,
	})
	if err != nil {
		if errors.Is(err, gateway_interfaces.ErrSelfToolNotFound) {
			return nil, apperrors.NewNotFoundError("self_tool_not_found", "self tool not found")
		}
		return nil, err
	}
	return responses.ToSelfToolResponse(payload), nil
}
