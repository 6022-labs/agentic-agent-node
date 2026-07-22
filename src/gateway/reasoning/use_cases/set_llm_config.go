package use_cases

import (
	"context"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
	"github.com/samber/lo"
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
		BaseUrl:  lo.EmptyableToPtr(strings.TrimSpace(request.BaseUrl)),
		ApiKey:   request.ApiKey,
	})
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
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
