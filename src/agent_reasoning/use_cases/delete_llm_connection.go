package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
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
func (u *DeleteLLMConnection) Execute(ctx context.Context) *usecase_result.Result {
	if err := u.repository.Delete(); err != nil {
		return usecase_result.InternalError("delete_failed", err.Error())
	}
	if err := u.secretStorage.Delete(ctx); err != nil {
		return usecase_result.InternalError("delete_failed", err.Error())
	}
	return usecase_result.NoContent()
}
