package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type CreateMcpConnection struct {
	client gateway_interfaces.MCPConfigClientInterface
}

func NewCreateMcpConnection(client gateway_interfaces.MCPConfigClientInterface) *CreateMcpConnection {
	return &CreateMcpConnection{client: client}
}

func (u *CreateMcpConnection) Execute(
	ctx context.Context,
	request *requests.CreateMcpConnectionRequest,
) (*responses.McpConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.client.Create(ctx, &models.ExternalCreateMCPConnectionPayload{
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
