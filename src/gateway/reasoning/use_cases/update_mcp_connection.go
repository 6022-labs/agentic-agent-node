package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type UpdateMcpConnection struct {
	client gateway_interfaces.MCPConfigClientInterface
}

func NewUpdateMcpConnection(client gateway_interfaces.MCPConfigClientInterface) *UpdateMcpConnection {
	return &UpdateMcpConnection{client: client}
}

func (u *UpdateMcpConnection) Execute(
	ctx context.Context,
	request *requests.UpdateMcpConnectionRequest,
) (*responses.McpConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.client.Update(ctx, request.Id, &models.ExternalUpdateMCPConnectionPayload{
		Name:                request.Name,
		Url:                 request.Url,
		AllowedToolPatterns: request.AllowedToolPatterns,
		HeaderAuth:          request.HeaderAuth,
	})
	if err != nil {
		return nil, err
	}
	return responses.ToMcpConnectionResponse(payload), nil
}
