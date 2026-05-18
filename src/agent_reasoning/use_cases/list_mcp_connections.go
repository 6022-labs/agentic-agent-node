package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
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
func (u *ListMCPConnections) Execute() *usecase_result.Result {
	connections, err := u.repository.List()
	if err != nil {
		return usecase_result.InternalError("list_failed", err.Error())
	}
	return usecase_result.Ok(responses.ToMCPConnectionListResponse(connections))
}
