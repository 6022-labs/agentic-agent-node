package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
)

type GetLLMConnection struct {
	repository repositories.LLMConnectionRepositoryInterface
}

func NewGetLLMConnection(repository repositories.LLMConnectionRepositoryInterface) *GetLLMConnection {
	return &GetLLMConnection{repository: repository}
}

// Execute returns the configured connection (without the api key — that lives
// in Vault and is only fetched at LLM-call time).
func (u *GetLLMConnection) Execute() *usecase_result.Result {
	connection, err := u.repository.Get()
	if err != nil {
		return usecase_result.InternalError("get_failed", err.Error())
	}
	if connection == nil {
		return usecase_result.NotFound("not_configured", "")
	}
	return usecase_result.Ok(responses.ToLLMConnectionResponse(connection))
}
