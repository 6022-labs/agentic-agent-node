package use_cases

import (
	"context"
	"fmt"

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
	*AbstractReasoningUseCase
	repository    repositories.MCPConnectionRepositoryInterface
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewUpdateMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *UpdateMCPConnection {
	return &UpdateMCPConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		repository:               repository,
		secretStorage:            secretStorage,
	}
}

// Execute applies full-replacement semantics: name, url, and the entire
// HeaderAuth map are overwritten. Header values absent from the new map are
// removed from Vault.
func (u *UpdateMCPConnection) Execute(ctx context.Context, id uuid.UUID, request *requests.UpdateMCPConnectionRequest) (*responses.MCPConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	existing, err := u.repository.GetById(ctx, agent.Id, id)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}
	if existing == nil {
		return nil, errors.NewNotFoundError("mcp_connection_not_found", "")
	}

	if err := u.dropRemovedHeaders(ctx, existing, request.HeaderAuth); err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}

	newKeys := models.MCPConnectionHeaderAuthKeys{}
	headerNames := make([]string, 0, len(request.HeaderAuth))
	for k := range request.HeaderAuth {
		headerNames = append(headerNames, k)
	}
	newKeys.FromArray(headerNames)

	existing.Name = request.Name
	existing.Url = request.Url
	existing.HeaderAuthKeys = newKeys
	updated, err := u.repository.Update(ctx, existing)
	if err != nil {
		return nil, errors.NewInternalError("update_failed", err.Error())
	}

	for k, v := range request.HeaderAuth {
		if err := u.secretStorage.SaveValue(ctx, updated.Id, k, v); err != nil {
			return nil, errors.NewInternalError("update_failed", fmt.Sprintf("failed to save header %q: %s", k, err.Error()))
		}
	}
	updated.HeaderAuth = request.HeaderAuth
	return responses.ToMCPConnectionResponse(updated), nil
}

func (u *UpdateMCPConnection) dropRemovedHeaders(ctx context.Context, existing *models.MCPConnection, next map[string]string) error {
	for _, key := range existing.HeaderAuthKeys {
		if _, kept := next[string(key)]; kept {
			continue
		}
		if err := u.secretStorage.DeleteValue(ctx, existing.Id, string(key)); err != nil {
			return err
		}
	}
	return nil
}
