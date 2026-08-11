package use_cases

import (
	"context"

	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
)

type SetupLLMConnection struct {
	*AbstractReasoningUseCase
	repository             repositories.LLMConnectionRepositoryInterface
	secretStorage          reasoning_interfaces.LLMSecretStorageInterface
	llmConnectionValidator reasoning_interfaces.LLMConnectionValidatorInterface
}

func NewSetupLLMConnection(
	repository repositories.LLMConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.LLMSecretStorageInterface,
	llmConnectionValidator reasoning_interfaces.LLMConnectionValidatorInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *SetupLLMConnection {
	return &SetupLLMConnection{
		AbstractReasoningUseCase: NewAbstractReasoningUseCase(agentRepository),
		repository:               repository,
		secretStorage:            secretStorage,
		llmConnectionValidator:   llmConnectionValidator,
	}
}

// Execute probes before Vault and resolves the api key before upserting, so the row never points at a missing secret.
func (u *SetupLLMConnection) Execute(
	ctx context.Context,
	request *requests.SetupLLMConnectionRequest,
) (*responses.LLMConnectionResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	if err := u.llmConnectionValidator.Validate(ctx, &models.LLMConnection{
		Platform: request.Platform,
		Model:    request.Model,
		BaseUrl:  request.BaseUrl,
	}, request.ApiKey); err != nil {
		return nil, err
	}

	if err := u.resolveApiKey(ctx, request.ApiKey); err != nil {
		return nil, err
	}

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	connection, err := u.repository.Upsert(ctx, &models.LLMConnection{
		AgentId:     agent.Id,
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

// resolveApiKey saves a provided key, otherwise keeps the stored one — required when none exists yet.
func (u *SetupLLMConnection) resolveApiKey(ctx context.Context, apiKey string) error {
	if apiKey != "" {
		if err := u.secretStorage.Save(ctx, apiKey); err != nil {
			return errors.NewInternalError("setup_failed", err.Error())
		}
		return nil
	}

	_, found, err := u.secretStorage.Get(ctx)
	if err != nil {
		return errors.NewInternalError("setup_failed", err.Error())
	}
	if found {
		return nil
	}

	return errors.NewValidationError("apiKey", "cannot be empty on first setup")
}
