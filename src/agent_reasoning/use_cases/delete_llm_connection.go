package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
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
	agentResolver identity_interfaces.AgentResolverInterface,
) *DeleteLLMConnection {
	return &DeleteLLMConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentResolver),
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
