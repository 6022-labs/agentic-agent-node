package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
)

type DeleteLLMConnection struct {
	repository    repositories.LLMConnectionRepositoryInterface
	secretStorage reasoning_interfaces.LLMSecretStorageInterface
}

func NewDeleteLLMConnection(
	repository repositories.LLMConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.LLMSecretStorageInterface,
) *DeleteLLMConnection {
	return &DeleteLLMConnection{repository: repository, secretStorage: secretStorage}
}

// Execute removes the row first, then the secret. Reverse would leave a row
// pointing at a missing secret if vault delete fails.
func (u *DeleteLLMConnection) Execute(ctx context.Context) error {
	if err := u.repository.Delete(); err != nil {
		return err
	}
	return u.secretStorage.Delete(ctx)
}
