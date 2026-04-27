package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/google/uuid"
)

type GetMCPConnection struct {
	repository    repositories.MCPConnectionRepositoryInterface
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewGetMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
) *GetMCPConnection {
	return &GetMCPConnection{repository: repository, secretStorage: secretStorage}
}

// Execute returns the connection with HeaderAuth hydrated from Vault. Returns
// (nil, nil) if no connection exists with that id.
func (u *GetMCPConnection) Execute(ctx context.Context, id uuid.UUID) (*models.MCPConnection, error) {
	connection, err := u.repository.GetById(id)
	if err != nil || connection == nil {
		return connection, err
	}
	connection.HeaderAuth = make(map[string]string, len(connection.HeaderAuthKeys))
	for _, key := range connection.HeaderAuthKeys {
		value, found, err := u.secretStorage.GetValue(ctx, connection.Id, string(key))
		if err != nil {
			return nil, err
		}
		if found {
			connection.HeaderAuth[string(key)] = value
		}
	}
	return connection, nil
}
