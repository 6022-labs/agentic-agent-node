package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SetupLLMConnection struct {
	repository    repositories.LLMConnectionRepositoryInterface
	secretStorage reasoning_interfaces.LLMSecretStorageInterface
}

func NewSetupLLMConnection(
	repository repositories.LLMConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.LLMSecretStorageInterface,
) *SetupLLMConnection {
	return &SetupLLMConnection{repository: repository, secretStorage: secretStorage}
}

// Execute persists the api key in Vault first, then upserts the row. If the
// upsert fails the secret stays in Vault — next setup attempt will overwrite
// it. The reverse order would risk a row pointing at a missing secret.
func (u *SetupLLMConnection) Execute(ctx context.Context, request *requests.SetupLLMConnectionRequest) (*responses.LLMConnectionResponse, error) {
	if err := u.secretStorage.Save(ctx, request.ApiKey); err != nil {
		return nil, errors.NewInternalError("setup_failed", err.Error())
	}
	connection, err := u.repository.Upsert(ctx, &models.LLMConnection{
		Platform:    request.Platform,
		Model:       request.Model,
		BaseUrl:     request.BaseUrl,
		SetupStatus: reasoning_enums.LLMConnectionSetupStatusEstablished,
	})
	if err != nil {
		return nil, errors.NewInternalError("setup_failed", err.Error())
	}
	return responses.ToLLMConnectionResponse(connection), nil
}
