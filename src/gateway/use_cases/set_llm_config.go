package use_cases

import (
	"context"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
)

type SetLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewSetLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *SetLlmConfig {
	return &SetLlmConfig{client: client}
}

func (u *SetLlmConfig) Execute(ctx context.Context, request *requests.SetLlmConfigRequest) (*responses.GetLlmConfigResponse, error) {
	if err := request.ValidateAndSanitize(); err != nil {
		return nil, err
	}

	payload, err := u.client.Setup(ctx, &models.ExternalSetupLLMConnectionPayload{
		Platform: request.Provider,
		Model:    request.Model,
		BaseUrl:  u.toBaseUrlPtr(request.BaseUrl),
		ApiKey:   request.ApiKey,
	})
	if err != nil {
		return nil, errors.NewUnavailableError("runtime_unavailable", err.Error())
	}
	return u.toResponse(payload), nil
}

func (u *SetLlmConfig) toBaseUrlPtr(s string) *string {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) == 0 {
		return nil
	}
	return &trimmed
}

func (u *SetLlmConfig) toResponse(p *models.ExternalLLMConnectionPayload) *responses.GetLlmConfigResponse {
	baseUrl := ""
	if p.BaseUrl != nil {
		baseUrl = *p.BaseUrl
	}
	return &responses.GetLlmConfigResponse{
		Id:          p.Id,
		Provider:    p.Platform,
		Model:       p.Model,
		BaseUrl:     baseUrl,
		SetupStatus: p.SetupStatus,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
