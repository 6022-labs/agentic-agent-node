package use_cases

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
)

type GetLLMConnection struct {
	repository repositories.LLMConnectionRepositoryInterface
}

func NewGetLLMConnection(repository repositories.LLMConnectionRepositoryInterface) *GetLLMConnection {
	return &GetLLMConnection{repository: repository}
}

// Execute returns the configured connection (without the api key — that lives
// in Vault and is only fetched at LLM-call time).
func (u *GetLLMConnection) Execute() (*models.LLMConnection, error) {
	return u.repository.Get()
}
