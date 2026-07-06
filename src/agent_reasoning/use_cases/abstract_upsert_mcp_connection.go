package use_cases

import (
	"context"
	"fmt"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// AbstractUpsertMCPConnection holds the persistence and secret storage shared by the create and update MCP-connection use cases.
type AbstractUpsertMCPConnection struct {
	*AbstractReasoningUseCase
	mcpConnectionRepository    repositories.MCPConnectionRepositoryInterface
	mcpHeaderAuthSecretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewAbstractUpsertMCPConnection(
	mcpConnectionRepository repositories.MCPConnectionRepositoryInterface,
	mcpHeaderAuthSecretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *AbstractUpsertMCPConnection {
	return &AbstractUpsertMCPConnection{
		AbstractReasoningUseCase:   NewAbstractReasoningUseCase(agentRepository),
		mcpConnectionRepository:    mcpConnectionRepository,
		mcpHeaderAuthSecretStorage: mcpHeaderAuthSecretStorage,
	}
}

// ensureNameAvailable keeps function names unambiguous: a duplicate name would collide in the tool list and dispatch.
func (u *AbstractUpsertMCPConnection) ensureNameAvailable(ctx context.Context, agentId uuid.UUID, name string, excludedId uuid.UUID) error {
	existing, err := u.mcpConnectionRepository.ListByAgentId(ctx, agentId)
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

// headerAuthKeys derives the persisted header-name set from the request; values stay in Vault.
func (u *AbstractUpsertMCPConnection) headerAuthKeys(request *requests.SaveMCPConnectionRequest) models.MCPConnectionHeaderAuthKeys {
	names := make([]string, 0, len(request.HeaderAuth))
	for k := range request.HeaderAuth {
		names = append(names, k)
	}
	keys := models.MCPConnectionHeaderAuthKeys{}
	keys.FromArray(names)
	return keys
}

func (u *AbstractUpsertMCPConnection) saveHeaderValues(ctx context.Context, connectionId uuid.UUID, headers map[string]string) error {
	for k, v := range headers {
		if err := u.mcpHeaderAuthSecretStorage.SaveValue(ctx, connectionId, k, v); err != nil {
			return fmt.Errorf("failed to save header %q: %w", k, err)
		}
	}
	return nil
}
