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
	"github.com/google/uuid"
)

type UpdateMCPConnection struct {
	*AbstractUpsertMCPConnection
}

func NewUpdateMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *UpdateMCPConnection {
	return &UpdateMCPConnection{
		AbstractUpsertMCPConnection: NewAbstractUpsertMCPConnection(repository, secretStorage, agentRepository),
	}
}

// Execute applies full-replacement semantics: name, url, and the entire
// HeaderAuth map are overwritten. Header values absent from the new map are
// removed from Vault.
func (u *UpdateMCPConnection) Execute(ctx context.Context, id uuid.UUID, request *requests.SaveMCPConnectionRequest) (*responses.MCPConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := u.mcpConnectionRepository.GetById(ctx, agent.Id, id)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if existing == nil {
		return nil, errors.NewNotFoundError("mcp_connection_not_found", "")
	}

	if request.Name != existing.Name {
		if err := u.ensureNameAvailable(ctx, agent.Id, request.Name, existing.Id); err != nil {
			return nil, err
		}
	}

	if err := u.dropRemovedHeaders(ctx, existing, request.HeaderAuth); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}

	existing.Name = request.Name
	existing.Url = request.Url
	existing.HeaderAuthKeys = u.headerAuthKeys(request)
	updated, err := u.mcpConnectionRepository.Update(ctx, existing)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}

	if err := u.saveHeaderValues(ctx, updated.Id, request.HeaderAuth); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	updated.HeaderAuth = request.HeaderAuth
	return responses.ToMCPConnectionResponse(updated), nil
}

func (u *UpdateMCPConnection) dropRemovedHeaders(ctx context.Context, existing *models.MCPConnection, next map[string]string) error {
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
