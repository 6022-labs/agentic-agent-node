package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ListMCPConnections struct {
	*AbstractReasoningUseCase
	repository repositories.MCPConnectionRepositoryInterface
}

func NewListMCPConnections(
	repository repositories.MCPConnectionRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *ListMCPConnections {
	return &ListMCPConnections{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		repository:               repository,
	}
}

// Execute returns every configured MCP connection. Header auth values are
// not hydrated — fetch a single connection via GetMCPConnection if you need
// the actual header values.
func (u *ListMCPConnections) Execute(ctx context.Context) (*responses.MCPConnectionListResponse, error) {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}
	connections, err := u.repository.ListByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("list_failed", err.Error())
	}
	return responses.ToMCPConnectionListResponse(connections), nil
}
