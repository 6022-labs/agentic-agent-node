package use_cases

import (
	"context"
	"errors"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/responses"
)

type GetLlmConfig struct {
	client gateway_interfaces.LLMConfigClientInterface
}

func NewGetLlmConfig(client gateway_interfaces.LLMConfigClientInterface) *GetLlmConfig {
	return &GetLlmConfig{client: client}
}

func (u *GetLlmConfig) Execute(ctx context.Context) (*responses.GetLlmConfigResponse, error) {
	payload, err := u.client.Get(ctx)
	if errors.Is(err, gateway_interfaces.ErrLLMConnectionNotFound) {
		return nil, apperrors.NewNotFoundError("not_configured", "")
	}
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *GetLlmConfig) toResponse(p *models.ExternalLLMConnectionPayload) *responses.GetLlmConfigResponse {
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
