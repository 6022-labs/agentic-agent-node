package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// AbstractUpsertMCPConnection holds the persistence and secret storage shared by the create and update MCP-connection use cases.
type AbstractUpsertMCPConnection struct {
	*AbstractReasoningUseCase
	repository    repositories.MCPConnectionRepositoryInterface
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewAbstractUpsertMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *AbstractUpsertMCPConnection {
	return &AbstractUpsertMCPConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		repository:               repository,
		secretStorage:            secretStorage,
	}
}

// ensureNameAvailable keeps function names unambiguous: a duplicate name would collide in the tool list and dispatch.
func (u *AbstractUpsertMCPConnection) ensureNameAvailable(ctx context.Context, agentId uuid.UUID, name string, excludedId uuid.UUID) error {
	existing, err := u.repository.ListByAgentId(ctx, agentId)
	if err != nil {
		return errors.NewInternalError("mcp_connection_lookup_failed", err.Error())
	}
	for _, connection := range existing {
		if connection.Id != excludedId && connection.Name == name {
			return errors.NewConflictError("mcp_connection_name_taken", "an MCP connection named \""+name+"\" already exists")
		}
	}
	return nil
}
