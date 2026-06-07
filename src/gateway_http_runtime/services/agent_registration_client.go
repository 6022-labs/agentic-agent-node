package services

import (
	"context"
	"errors"
	"net/http"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

const (
	agentRegistrationPath          = "/internal/agent/registration"
	agentRegistrationMintInfoPath  = "/internal/agent/registration/mint-info"
	agentRegistrationMintCheckPath = "/internal/agent/registration/mint/check"
	agentRegistrationEnsInfoPath   = "/internal/agent/registration/ens-info"
	agentRegistrationEnsCheckPath  = "/internal/agent/registration/ens/check"
	agentRegistrationSelfPath      = "/internal/agent/registration/self"
)

type HttpAgentRegistrationClient struct {
	client *clients.RuntimeHttpClient
}

func NewHttpAgentRegistrationClient(client *clients.RuntimeHttpClient) *HttpAgentRegistrationClient {
	return &HttpAgentRegistrationClient{client: client}
}

func (c *HttpAgentRegistrationClient) Save(ctx context.Context, body *runtime_models.ExternalSaveAgentRegistrationPayload) (*runtime_models.ExternalAgentRegistrationPayload, error) {
	var payload runtime_models.ExternalAgentRegistrationPayload
	if err := c.client.Post(ctx, agentRegistrationPath, body, &payload); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) Get(ctx context.Context) (*runtime_models.ExternalAgentRegistrationPayload, error) {
	var payload runtime_models.ExternalAgentRegistrationPayload
	if err := c.client.Get(ctx, agentRegistrationPath, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) MintInfo(ctx context.Context) (*runtime_models.ExternalMintInfoPayload, error) {
	var payload runtime_models.ExternalMintInfoPayload
	if err := c.client.Get(ctx, agentRegistrationMintInfoPath, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) CheckMint(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error) {
	var payload runtime_models.ExternalRegistrationStatusPayload
	if err := c.client.Post(ctx, agentRegistrationMintCheckPath, nil, &payload); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) EnsInfo(ctx context.Context) (*runtime_models.ExternalEnsInfoPayload, error) {
	var payload runtime_models.ExternalEnsInfoPayload
	if err := c.client.Get(ctx, agentRegistrationEnsInfoPath, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) CheckEns(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error) {
	var payload runtime_models.ExternalRegistrationStatusPayload
	if err := c.client.Post(ctx, agentRegistrationEnsCheckPath, nil, &payload); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) Self(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error) {
	var payload runtime_models.ExternalRegistrationStatusPayload
	if err := c.client.Post(ctx, agentRegistrationSelfPath, nil, &payload); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

// mapRuntimeError maps the runtime's HTTP-status error to a domain error,
// preserving 400/404/409 and collapsing the rest to 502.
func mapRuntimeError(err error) error {
	var statusErr *clients.RuntimeStatusError
	if errors.As(err, &statusErr) {
		switch statusErr.StatusCode {
		case http.StatusBadRequest:
			return apperrors.NewValidationError(statusErr.Field, statusErr.Message)
		case http.StatusNotFound:
			return apperrors.NewNotFoundError(statusErr.Code, statusErr.Message)
		case http.StatusConflict:
			return apperrors.NewConflictError(statusErr.Code, statusErr.Message)
		}
	}
	return apperrors.NewUnavailableError("runtime_unavailable", err.Error())
}
