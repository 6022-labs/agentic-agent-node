package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type GetLLMConnection struct {
	*AbstractReasoningUseCase
	repository repositories.LLMConnectionRepositoryInterface
}

func NewGetLLMConnection(
	repository repositories.LLMConnectionRepositoryInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *GetLLMConnection {
	return &GetLLMConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		repository:               repository,
	}
}

// Execute returns the configured connection (without the api key — that lives
// in Vault and is only fetched at LLM-call time).
func (u *GetLLMConnection) Execute(ctx context.Context) (*responses.LLMConnectionResponse, error) {
	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}
	connection, err := u.repository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("get_failed", err)
	}
	if connection == nil {
		return nil, errors.NewNotFoundError("not_configured", "")
	}
	return responses.ToLLMConnectionResponse(connection), nil
}
