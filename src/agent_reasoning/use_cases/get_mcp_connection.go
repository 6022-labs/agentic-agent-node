package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
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

// Execute returns the connection with HeaderAuth hydrated from Vault.
func (u *GetMCPConnection) Execute(ctx context.Context, id uuid.UUID) *usecase_result.Result {
	connection, err := u.repository.GetById(ctx, id)
	if err != nil {
		return usecase_result.InternalError("get_failed", err.Error())
	}
	if connection == nil {
		return usecase_result.NotFound("mcp_connection_not_found", "")
	}

	connection.HeaderAuth = make(map[string]string, len(connection.HeaderAuthKeys))
	for _, key := range connection.HeaderAuthKeys {
		value, found, err := u.secretStorage.GetValue(ctx, connection.Id, string(key))
		if err != nil {
			return usecase_result.InternalError("get_failed", err.Error())
		}
		if found {
			connection.HeaderAuth[string(key)] = value
		}
	}
	return usecase_result.Ok(responses.ToMCPConnectionResponse(connection))
}
