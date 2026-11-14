package use_cases

import (
	"context"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// AbstractUpsertMCPConnection holds the behavior shared by the create and update MCP-connection use cases.
type AbstractUpsertMCPConnection struct {
	*AbstractMCPConnectionUseCase
}

func NewAbstractUpsertMCPConnection(
	mcpConnectionRepository repositories.MCPConnectionRepositoryInterface,
	mcpHeaderAuthSecretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	mcpRequester reasoning_interfaces.McpRequesterInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *AbstractUpsertMCPConnection {
	return &AbstractUpsertMCPConnection{
		AbstractMCPConnectionUseCase: NewAbstractMCPConnectionUseCase(
			mcpConnectionRepository,
			mcpHeaderAuthSecretStorage,
			mcpRequester,
			agentResolver,
		),
	}
}

// probeListTools rejects a save whose server cannot list its tools.
func (u *AbstractUpsertMCPConnection) probeListTools(
	ctx context.Context,
	name, url string,
	headers map[string]string,
) error {
	_, err := u.listTools(ctx, &models.MCPConnection{Name: name, Url: url, HeaderAuth: headers})
	return err
}

// ensureNameAvailable keeps function names unambiguous: a duplicate name would collide in the tool list and dispatch.
func (u *AbstractUpsertMCPConnection) ensureNameAvailable(
	ctx context.Context,
	agentId uuid.UUID,
	name string,
	excludedId uuid.UUID,
) error {
	existing, err := u.mcpConnectionRepository.ListByAgentId(ctx, agentId)
	if err != nil {
		return errors.NewInternalError("mcp_connection_lookup_failed", err)
	}
	for _, connection := range existing {
		if connection.Id != excludedId && connection.Name == name {
			return errors.NewConflictError(
				"mcp_connection_name_taken",
				"an MCP connection named \""+name+"\" already exists",
			)
		}
	}
	return nil
}

// headerAuthKeys derives the persisted header-name set from the request; values stay in Vault.
func (u *AbstractUpsertMCPConnection) headerAuthKeys(
	request *requests.MCPHeaderAuthRequest,
) models.MCPConnectionHeaderAuthKeys {
	names := make([]string, 0, len(request.HeaderAuth))
	for k := range request.HeaderAuth {
		names = append(names, k)
	}
	keys := models.MCPConnectionHeaderAuthKeys{}
	keys.FromArray(names)
	return keys
}

func (u *AbstractUpsertMCPConnection) saveHeaderValues(
	ctx context.Context,
	connectionId uuid.UUID,
	headers map[string]string,
) error {
	for k, v := range headers {
		if err := u.mcpHeaderAuthSecretStorage.SaveValue(ctx, connectionId, k, v); err != nil {
			return fmt.Errorf("failed to save header %q: %w", k, err)
		}
	}
	return nil
}
