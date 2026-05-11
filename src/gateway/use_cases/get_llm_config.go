package use_cases

import (
	"context"
	"errors"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

// ErrLLMConnectionNotConfigured is returned by GetLlmConfig and DeleteLlmConfig
// when the runtime has no LLM connection saved.
var ErrLLMConnectionNotConfigured = errors.New("llm connection not configured")

type GetLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewGetLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *GetLlmConfig {
	return &GetLlmConfig{client: client}
}

func (u *GetLlmConfig) Execute(ctx context.Context) (*responses.GetLlmConfigResponse, error) {
	payload, err := u.client.Get(ctx)
	if errors.Is(err, gateway_interfaces.ErrLLMConnectionNotFound) {
		return nil, ErrLLMConnectionNotConfigured
	}
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
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
