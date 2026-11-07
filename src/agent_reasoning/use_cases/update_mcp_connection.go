package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type UpdateMCPConnection struct {
	*AbstractUpsertMCPConnection
}

func NewUpdateMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	mcpRequester reasoning_interfaces.McpRequesterInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *UpdateMCPConnection {
	return &UpdateMCPConnection{
		AbstractUpsertMCPConnection: NewAbstractUpsertMCPConnection(
			repository,
			secretStorage,
			mcpRequester,
			agentRepository,
		),
	}
}

// Execute applies full-replacement semantics: name, url, and the entire
// HeaderAuth map are overwritten, and dropped headers leave Vault.
func (u *UpdateMCPConnection) Execute(
	ctx context.Context,
	request *requests.UpdateMCPConnectionRequest,
) (*responses.MCPConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := u.mcpConnectionRepository.GetById(ctx, agent.Id, request.Id)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err)
	}
	if existing == nil {
		return nil, errors.NewNotFoundError("mcp_connection_not_found", "")
	}

	if request.Name != existing.Name {
		if err := u.ensureNameAvailable(ctx, agent.Id, request.Name, existing.Id); err != nil {
			return nil, err
		}
	}

	headers, err := u.resolveHeaderValues(ctx, existing, request.HeaderAuth)
	if err != nil {
		return nil, err
	}

	if err := u.probeListTools(ctx, request.Name, request.Url, headers); err != nil {
		return nil, err
	}

	if err := u.dropRemovedHeaders(ctx, existing, request.HeaderAuth); err != nil {
		return nil, errors.NewInternalError("update_failed", err)
	}

	existing.Name = request.Name
	existing.Url = request.Url
	existing.AllowedToolPatterns = request.AllowedToolPatterns
	existing.HeaderAuthKeys = u.headerAuthKeys(&request.MCPHeaderAuthRequest)
	updated, err := u.mcpConnectionRepository.Update(ctx, existing)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err)
	}

	if err := u.saveHeaderValues(ctx, updated.Id, headers); err != nil {
		return nil, errors.NewInternalError("update_failed", err)
	}
	updated.HeaderAuth = headers
	return responses.ToMCPConnectionResponse(updated), nil
}

// resolveHeaderValues fills blank values from Vault: the console never holds them, so blank means "unchanged".
func (u *UpdateMCPConnection) resolveHeaderValues(
	ctx context.Context,
	existing *models.MCPConnection,
	requested map[string]string,
) (map[string]string, error) {
	resolved := make(map[string]string, len(requested))
	for key, value := range requested {
		if value != "" {
			resolved[key] = value
			continue
		}
		stored, found, err := u.mcpHeaderAuthSecretStorage.GetValue(ctx, existing.Id, key)
		if err != nil {
			return nil, errors.NewInternalError("update_failed", err)
		}
		if !found {
			return nil, errors.NewValidationError("headerAuth", "value for "+key+" cannot be empty")
		}
		resolved[key] = stored
	}
	return resolved, nil
}

func (u *UpdateMCPConnection) dropRemovedHeaders(
	ctx context.Context,
	existing *models.MCPConnection,
	next map[string]string,
) error {
	for _, key := range existing.HeaderAuthKeys {
		if _, kept := next[string(key)]; kept {
			continue
		}
		if err := u.mcpHeaderAuthSecretStorage.DeleteValue(ctx, existing.Id, string(key)); err != nil {
			return err
		}
	}
	return nil
}
