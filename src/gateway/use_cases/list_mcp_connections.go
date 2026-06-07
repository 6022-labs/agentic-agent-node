package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

type ListMcpConnections struct {
	client gateway_interfaces.MCPConfigClientInterface
}

func NewListMcpConnections(client gateway_interfaces.MCPConfigClientInterface) *ListMcpConnections {
	return &ListMcpConnections{client: client}
}

func (u *ListMcpConnections) Execute(ctx context.Context) (*responses.ListMcpConnectionsResponse, error) {
	payload, err := u.client.List(ctx)
	if err != nil {
		return nil, errors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	out := make([]responses.McpConnectionResponse, 0, len(payload.Connections))
	for i := range payload.Connections {
		out = append(out, *u.toResponse(&payload.Connections[i]))
	}
	return &responses.ListMcpConnectionsResponse{Connections: out}, nil
}

func (u *ListMcpConnections) toResponse(p *models.ExternalMCPConnectionPayload) *responses.McpConnectionResponse {
	return &responses.McpConnectionResponse{
		Id:             p.Id,
		Name:           p.Name,
		Url:            p.Url,
		HeaderAuthKeys: p.HeaderAuthKeys,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}
