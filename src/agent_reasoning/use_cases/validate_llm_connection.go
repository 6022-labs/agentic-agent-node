package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type ValidateLLMConnection struct {
	*AbstractLLMConnectionUseCase
}

func NewValidateLLMConnection(
	repository repositories.LLMConnectionRepositoryInterface,
	llmConnectionValidator reasoning_interfaces.LLMConnectionValidatorInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *ValidateLLMConnection {
	return &ValidateLLMConnection{
		AbstractLLMConnectionUseCase: NewAbstractLLMConnectionUseCase(
			repository,
			llmConnectionValidator,
			agentRepository,
		),
	}
}

// Execute probes the candidate connection without persisting anything.
func (u *ValidateLLMConnection) Execute(ctx context.Context, request *requests.SetupLLMConnectionRequest) error {
	if err := request.ValidateAndSanitize(); err != nil {
		return err
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return err
	}
	current, err := u.llmConnectionRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return errors.NewInternalError("validate_failed", err.Error())
	}
	return u.validateConnection(ctx, request, current)
}
