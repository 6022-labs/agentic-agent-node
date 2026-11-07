package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type DeleteLLMConnection struct {
	*AbstractReasoningUseCase
	repository    repositories.LLMConnectionRepositoryInterface
	secretStorage reasoning_interfaces.LLMSecretStorageInterface
}

func NewDeleteLLMConnection(
	repository repositories.LLMConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.LLMSecretStorageInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *DeleteLLMConnection {
	return &DeleteLLMConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		repository:               repository,
		secretStorage:            secretStorage,
	}
}

// Execute removes the row first, then the secret. Reverse would leave a row
// pointing at a missing secret if vault delete fails.
func (u *DeleteLLMConnection) Execute(ctx context.Context) error {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return err
	}
	if err := u.repository.DeleteByAgentId(ctx, agent.Id); err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	if err := u.secretStorage.Delete(ctx); err != nil {
		return errors.NewInternalError("delete_failed", err)
	}
	return nil
}
