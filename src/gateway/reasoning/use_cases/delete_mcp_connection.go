package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
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
	if err := u.client.Delete(ctx, request.Id); err != nil {
		return errors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	return nil
}
