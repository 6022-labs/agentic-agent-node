package use_cases

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/common/usecase_result"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

// ErrLLMConnectionNotConfigured is exposed for unit tests that assert on the
// semantic reason; the wire-side outcome is NotFound.
var ErrLLMConnectionNotConfigured = errors.New("llm connection not configured")

type GetLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewGetLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *GetLlmConfig {
	return &GetLlmConfig{client: client}
}

func (u *GetLlmConfig) Execute(ctx context.Context) *usecase_result.Result {
	payload, err := u.client.Get(ctx)
	if errors.Is(err, gateway_interfaces.ErrLLMConnectionNotFound) {
		return usecase_result.NotFound("not_configured", "")
	}
	if err != nil {
		return usecase_result.Unavailable("runtime_unavailable", err.Error())
	}
	return usecase_result.Ok(u.toResponse(payload))
}

func (u *GetLlmConfig) toResponse(p *runtime_models.ExternalLLMConnectionPayload) *responses.GetLlmConfigResponse {
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
