package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
)

type DeleteMcpConnection struct {
	client gateway_interfaces.MCPConfigClientInterface
}

func NewDeleteMcpConnection(client gateway_interfaces.MCPConfigClientInterface) *DeleteMcpConnection {
	return &DeleteMcpConnection{client: client}
}

func (u *DeleteMcpConnection) Execute(ctx context.Context, request *requests.DeleteMcpConnectionRequest) error {
	return u.client.Delete(ctx, request.Id)
}
