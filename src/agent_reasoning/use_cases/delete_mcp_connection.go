package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	"github.com/google/uuid"
)

type DeleteMCPConnection struct {
	repository    repositories.MCPConnectionRepositoryInterface
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewDeleteMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
) *DeleteMCPConnection {
	return &DeleteMCPConnection{repository: repository, secretStorage: secretStorage}
}

// Execute clears all header secrets from Vault first (the row knows their
// keys), then drops the row itself. No-op if the id doesn't exist.
func (u *DeleteMCPConnection) Execute(ctx context.Context, id uuid.UUID) *usecase_result.Result {
	existing, err := u.repository.GetById(id)
	if err != nil {
		return usecase_result.InternalError("delete_failed", err.Error())
	}
	if existing == nil {
		return usecase_result.NoContent()
	}
	for _, key := range existing.HeaderAuthKeys {
		if err := u.secretStorage.DeleteValue(ctx, existing.Id, string(key)); err != nil {
			return usecase_result.InternalError("delete_failed", err.Error())
		}
	}
	if err := u.repository.Delete(existing.Id); err != nil {
		return usecase_result.InternalError("delete_failed", err.Error())
	}
	return usecase_result.NoContent()
}
