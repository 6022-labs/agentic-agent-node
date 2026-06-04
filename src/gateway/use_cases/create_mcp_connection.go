package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type CreateMcpConnection struct {
	client gateway_interfaces.MCPConfigClientInterface
}

func NewCreateMcpConnection(client gateway_interfaces.MCPConfigClientInterface) *CreateMcpConnection {
	return &CreateMcpConnection{client: client}
}

func (u *CreateMcpConnection) Execute(ctx context.Context, request *requests.CreateMcpConnectionRequest) (*responses.McpConnectionResponse, error) {
	payload, err := u.client.Create(ctx, &runtime_models.ExternalCreateMCPConnectionPayload{
		Name:       request.Name,
		Url:        request.Url,
		HeaderAuth: u.toHeaderAuthMap(request),
	})
	if err != nil {
		return nil, errors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	return u.toResponse(payload), nil
}

func (u *CreateMcpConnection) toHeaderAuthMap(request *requests.CreateMcpConnectionRequest) map[string]string {
	if len(request.AuthHeaderName) == 0 {
		return nil
	}
	return map[string]string{request.AuthHeaderName: request.AuthHeaderValue}
}

func (u *CreateMcpConnection) toResponse(p *runtime_models.ExternalMCPConnectionPayload) *responses.McpConnectionResponse {
	return &responses.McpConnectionResponse{
		Id:             p.Id,
		Name:           p.Name,
		Url:            p.Url,
		HeaderAuthKeys: p.HeaderAuthKeys,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
