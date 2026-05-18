package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/google/uuid"
)

type DeleteMcpConnection struct {
	client gateway_interfaces.MCPConfigClientInterface
}

func NewDeleteMcpConnection(client gateway_interfaces.MCPConfigClientInterface) *DeleteMcpConnection {
	return &DeleteMcpConnection{client: client}
}

func (u *DeleteMcpConnection) Execute(ctx context.Context, id uuid.UUID) *usecase_result.Result {
	if err := u.client.Delete(ctx, id); err != nil {
		return usecase_result.Unavailable("runtime_unavailable", err.Error())
	}
	return usecase_result.NoContent()
}
