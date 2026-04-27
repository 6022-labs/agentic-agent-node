package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
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
func (u *ListMCPConnections) Execute() ([]models.MCPConnection, error) {
	return u.repository.List()
}
