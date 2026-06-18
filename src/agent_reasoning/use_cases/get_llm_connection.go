package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type GetLLMConnection struct {
	repository repositories.LLMConnectionRepositoryInterface
}

func NewGetLLMConnection(repository repositories.LLMConnectionRepositoryInterface) *GetLLMConnection {
	return &GetLLMConnection{repository: repository}
}

// Execute returns the configured connection (without the api key — that lives
// in Vault and is only fetched at LLM-call time).
func (u *GetLLMConnection) Execute(ctx context.Context) (*responses.LLMConnectionResponse, error) {
	connection, err := u.repository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("get_failed", err.Error())
	}
	if connection == nil {
		return nil, errors.NewNotFoundError("not_configured", "")
	}
	return responses.ToLLMConnectionResponse(connection), nil
}
