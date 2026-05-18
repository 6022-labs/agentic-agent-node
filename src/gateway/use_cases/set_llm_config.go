package use_cases

import (
	"context"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type SetLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewSetLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *SetLlmConfig {
	return &SetLlmConfig{client: client}
}

func (u *SetLlmConfig) Execute(ctx context.Context, request *requests.SetLlmConfigRequest) *usecase_result.Result {
	payload, err := u.client.Setup(ctx, &runtime_models.ExternalSetupLLMConnectionPayload{
		Platform: request.Provider,
		Model:    request.Model,
		BaseUrl:  u.toBaseUrlPtr(request.BaseUrl),
		ApiKey:   request.ApiKey,
	})
	if err != nil {
		return usecase_result.Unavailable("runtime_unavailable", err.Error())
	}
	return usecase_result.Ok(u.toResponse(payload))
}

func (u *SetLlmConfig) toBaseUrlPtr(s string) *string {
	trimmed := strings.TrimSpace(s)
	if len(trimmed) == 0 {
		return nil
	}
	return &trimmed
}

func (u *SetLlmConfig) toResponse(p *runtime_models.ExternalLLMConnectionPayload) *responses.GetLlmConfigResponse {
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
