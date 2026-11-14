package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeleteMCPConnection struct {
	*AbstractReasoningUseCase
	repository    repositories.MCPConnectionRepositoryInterface
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface
}

func NewDeleteMCPConnection(
	repository repositories.MCPConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.MCPHeaderAuthSecretStorageInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *DeleteMCPConnection {
	return &DeleteMCPConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentResolver),
		repository:               repository,
		secretStorage:            secretStorage,
	}
}

// Execute clears all header secrets from Vault first (the row knows their
// keys), then drops the row itself. No-op if the id doesn't exist.
func (u *DeleteMCPConnection) Execute(ctx context.Context, request *requests.DeleteMCPConnectionRequest) error {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return err
	}
	existing, err := u.repository.GetById(ctx, agent.Id, request.Id)
	if err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	if existing == nil {
		return nil
	}
	for _, key := range existing.HeaderAuthKeys {
		if err := u.secretStorage.DeleteValue(ctx, existing.Id, string(key)); err != nil {
			return errors.NewInternalError("delete_failed", err)
		}
	}
	if err := u.repository.DeleteById(ctx, agent.Id, existing.Id); err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	return nil
}
