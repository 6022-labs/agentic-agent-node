package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
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
		out = append(out, *responses.ToMcpConnectionResponse(&payload.Connections[i]))
	}
	return &responses.ListMcpConnectionsResponse{Connections: out}, nil
}
