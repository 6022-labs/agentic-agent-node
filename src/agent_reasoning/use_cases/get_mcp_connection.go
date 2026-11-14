package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type GetMCPConnection struct {
	*AbstractReasoningUseCase
	repository    repositories.MCPConnectionRepositoryInterface
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewGetMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *GetMCPConnection {
	return &GetMCPConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentResolver),
		repository:               repository,
		secretStorage:            secretStorage,
	}
}

// Execute returns the connection with HeaderAuth hydrated from Vault.
func (u *GetMCPConnection) Execute(
	ctx context.Context,
	request *requests.GetMCPConnectionRequest,
) (*responses.MCPConnectionResponse, error) {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}
	connection, err := u.repository.GetById(ctx, agent.Id, request.Id)
	if err != nil {
		return nil, errors.NewInternalError("get_failed", err)
	}
	if connection == nil {
		return nil, errors.NewNotFoundError("mcp_connection_not_found", "")
	}

	connection.HeaderAuth = make(map[string]string, len(connection.HeaderAuthKeys))
	for _, key := range connection.HeaderAuthKeys {
		value, found, err := u.secretStorage.GetValue(ctx, connection.Id, string(key))
		if err != nil {
			return nil, errors.NewInternalError("get_failed", err)
		}
		if found {
			connection.HeaderAuth[string(key)] = value
		}
	}
	return responses.ToMCPConnectionResponse(connection), nil
}
