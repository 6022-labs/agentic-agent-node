package use_cases

import (
	"context"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/samber/lo"
)

type ValidateLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewValidateLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *ValidateLlmConfig {
	return &ValidateLlmConfig{client: client}
}

func (u *ValidateLlmConfig) Execute(ctx context.Context, request *requests.SetLlmConfigRequest) error {
	if err := request.ValidateAndSanitize(); err != nil {
		return err
	}

	return u.client.Validate(ctx, &models.ExternalSetupLLMConnectionPayload{
		Platform: request.Provider,
		Model:    request.Model,
		BaseUrl:  lo.EmptyableToPtr(strings.TrimSpace(request.BaseUrl)),
		ApiKey:   request.ApiKey,
	})
}
