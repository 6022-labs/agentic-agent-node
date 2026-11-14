package use_cases

import (
	"context"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

// AbstractLLMConnectionUseCase holds the api key rule and probe shared by the setup and validate use cases.
type AbstractLLMConnectionUseCase struct {
	*AbstractReasoningUseCase
	llmConnectionRepository repositories.LLMConnectionRepositoryInterface
	llmConnectionValidator  reasoning_interfaces.LLMConnectionValidatorInterface
}

func NewAbstractLLMConnectionUseCase(
	llmConnectionRepository repositories.LLMConnectionRepositoryInterface,
	llmConnectionValidator reasoning_interfaces.LLMConnectionValidatorInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *AbstractLLMConnectionUseCase {
	return &AbstractLLMConnectionUseCase{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentResolver),
		llmConnectionRepository:  llmConnectionRepository,
		llmConnectionValidator:   llmConnectionValidator,
	}
}

// validateConnection applies the api key rule against current, then probes the candidate connection.
func (u *AbstractLLMConnectionUseCase) validateConnection(
	ctx context.Context,
	request *requests.SetupLLMConnectionRequest,
	current *models.LLMConnection,
) error {
	if err := u.validateApiKey(request, current); err != nil {
		return err
	}
	return u.llmConnectionValidator.Validate(ctx, &models.LLMConnection{
		Platform: request.Platform,
		Model:    request.Model,
		BaseUrl:  request.BaseUrl,
	}, request.ApiKey)
}

// validateApiKey rejects a key on a local provider and requires one when choosing a new cloud provider.
func (u *AbstractLLMConnectionUseCase) validateApiKey(
	request *requests.SetupLLMConnectionRequest,
	current *models.LLMConnection,
) error {
	if u.usesLocalPlatform(request.Platform) {
		if request.ApiKey != "" {
			return errors.NewValidationError("apiKey", "must be empty for a local provider")
		}
		return nil
	}
	if request.ApiKey != "" || (current != nil && current.Platform == request.Platform) {
		return nil
	}
	return errors.NewValidationError("apiKey", "is required when choosing a new provider")
}

// usesLocalPlatform reports whether the provider runs keyless on the agent's own network.
func (u *AbstractLLMConnectionUseCase) usesLocalPlatform(platform reasoning_enums.LLMConnectionPlatform) bool {
	return reasoning_enums.GetAllLocalLLMConnectionPlatforms().Contains(platform)
}
