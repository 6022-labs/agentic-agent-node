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
	*AbstractLLMConnectionUseCase
	secretStorage reasoning_interfaces.LLMSecretStorageInterface
}

func NewSetupLLMConnection(
	repository repositories.LLMConnectionRepositoryInterface,
	secretStorage reasoning_interfaces.LLMSecretStorageInterface,
	llmConnectionValidator reasoning_interfaces.LLMConnectionValidatorInterface,
	agentRepository identity_repositories.AgentRepositoryInterface,
) *SetupLLMConnection {
	return &SetupLLMConnection{
		AbstractLLMConnectionUseCase: NewAbstractLLMConnectionUseCase(
			repository,
			llmConnectionValidator,
			agentRepository,
		),
		secretStorage: secretStorage,
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

	agent, err := u.resolveAgent(ctx)
	if err != nil {
		return nil, err
	}

	current, err := u.llmConnectionRepository.GetByAgentId(ctx, agent.Id)
	if err != nil {
		return nil, errors.NewInternalError("setup_failed", err.Error())
	}
	if err := u.validateConnection(ctx, request, current); err != nil {
		return nil, err
	}

	if err := u.resolveApiKey(ctx, request); err != nil {
		return nil, err
	}

	connection, err := u.llmConnectionRepository.Upsert(ctx, &models.LLMConnection{
		AgentId:     agent.Id,
		Platform:    request.Platform,
		Model:       request.Model,
		BaseUrl:     request.BaseUrl,
		SetupStatus: reasoning_enums.LLMConnectionSetupStatusEstablished,
	})
	if err != nil {
		return nil, errors.NewInternalError("setup_failed", err.Error())
	}

	// After the upsert, on every local save, so a failed delete heals on retry.
	if u.usesLocalPlatform(request.Platform) {
		if err := u.secretStorage.Delete(ctx); err != nil {
			return nil, errors.NewInternalError("setup_failed", err.Error())
		}
	}
	return responses.ToLLMConnectionResponse(connection), nil
}

// resolveApiKey saves a given key; a keyless local provider needs none, else a key must already be stored.
func (u *SetupLLMConnection) resolveApiKey(ctx context.Context, request *requests.SetupLLMConnectionRequest) error {
	if request.ApiKey != "" {
		if err := u.secretStorage.Save(ctx, request.ApiKey); err != nil {
			return errors.NewInternalError("setup_failed", err.Error())
		}
		return nil
	}
	if u.usesLocalPlatform(request.Platform) {
		return nil
	}

	_, found, err := u.secretStorage.Get(ctx)
	if err != nil {
		return errors.NewInternalError("setup_failed", err.Error())
	}
	if !found {
		return errors.NewValidationError("apiKey", "cannot be empty on first setup")
	}
	return nil
}
