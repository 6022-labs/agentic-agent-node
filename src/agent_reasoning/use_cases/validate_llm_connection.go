package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
)

type ValidateLLMConnection struct {
	llmConnectionValidator reasoning_interfaces.LLMConnectionValidatorInterface
}

func NewValidateLLMConnection(
	llmConnectionValidator reasoning_interfaces.LLMConnectionValidatorInterface,
) *ValidateLLMConnection {
	return &ValidateLLMConnection{llmConnectionValidator: llmConnectionValidator}
}

// Execute probes the candidate connection without persisting anything.
func (u *ValidateLLMConnection) Execute(ctx context.Context, request *requests.SetupLLMConnectionRequest) error {
	if err := request.ValidateAndSanitize(); err != nil {
		return err
	}

	return u.llmConnectionValidator.Validate(ctx, &models.LLMConnection{
		Platform: request.Platform,
		Model:    request.Model,
		BaseUrl:  request.BaseUrl,
	}, request.ApiKey)
}
