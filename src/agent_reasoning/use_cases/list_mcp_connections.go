package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListMCPConnections struct {
	repository repositories.MCPConnectionRepositoryInterface
}

func NewListMCPConnections(repository repositories.MCPConnectionRepositoryInterface) *ListMCPConnections {
	return &ListMCPConnections{repository: repository}
}

// Execute returns every configured MCP connection. Header auth values are
// not hydrated — fetch a single connection via GetMCPConnection if you need
// the actual header values.
func (u *ListMCPConnections) Execute(ctx context.Context) (*responses.MCPConnectionListResponse, error) {
	connections, err := u.repository.List(ctx)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}
	return responses.ToMCPConnectionListResponse(connections), nil
}
