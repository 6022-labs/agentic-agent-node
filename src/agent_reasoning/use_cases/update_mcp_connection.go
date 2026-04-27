package use_cases

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/google/uuid"
)

type UpdateMCPConnection struct {
	repository    repositories.MCPConnectionRepositoryInterface
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewUpdateMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
) *UpdateMCPConnection {
	return &UpdateMCPConnection{repository: repository, secretStorage: secretStorage}
}

// Execute applies full-replacement semantics: name, url, and the entire
// HeaderAuth map are overwritten. Header values absent from the new map are
// removed from Vault. Returns (nil, nil) if no connection exists with that id.
func (u *UpdateMCPConnection) Execute(ctx context.Context, id uuid.UUID, request *requests.UpdateMCPConnectionRequest) (*models.MCPConnection, error) {
	existing, err := u.repository.GetById(id)
	if err != nil || existing == nil {
		return existing, err
	}

	if err := u.dropRemovedHeaders(ctx, existing, request.HeaderAuth); err != nil {
		return nil, err
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
	updated, err := u.repository.Update(existing)
	if err != nil {
		return nil, err
	}

	for k, v := range request.HeaderAuth {
		if err := u.secretStorage.SaveValue(ctx, updated.Id, k, v); err != nil {
			return nil, fmt.Errorf("failed to save header %q: %w", k, err)
		}
	}
	updated.HeaderAuth = request.HeaderAuth
	return updated, nil
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
