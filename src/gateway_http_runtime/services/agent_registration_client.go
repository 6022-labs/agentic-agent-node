package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type HttpAgentRegistrationClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAgentRegistrationClient(client clients.ClientWithResponsesInterface) *HttpAgentRegistrationClient {
	return &HttpAgentRegistrationClient{client: client}
}

func (c *HttpAgentRegistrationClient) Save(ctx context.Context, body *runtime_models.ExternalSaveAgentRegistrationPayload) (*runtime_models.ExternalAgentRegistrationPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	resp, err := c.client.PostInternalAgentRegistrationWithBodyWithResponse(ctx, jsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload runtime_models.ExternalAgentRegistrationPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) Get(ctx context.Context) (*runtime_models.ExternalAgentRegistrationPayload, error) {
	resp, err := c.client.GetInternalAgentRegistrationWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload runtime_models.ExternalAgentRegistrationPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) MintInfo(ctx context.Context) (*runtime_models.ExternalMintInfoPayload, error) {
	resp, err := c.client.GetInternalAgentRegistrationMintInfoWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload runtime_models.ExternalMintInfoPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) CheckMint(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error) {
	resp, err := c.client.PostInternalAgentRegistrationMintCheckWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload runtime_models.ExternalRegistrationStatusPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) EnsInfo(ctx context.Context) (*runtime_models.ExternalEnsInfoPayload, error) {
	resp, err := c.client.GetInternalAgentRegistrationEnsInfoWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload runtime_models.ExternalEnsInfoPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) CheckEns(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error) {
	resp, err := c.client.PostInternalAgentRegistrationEnsCheckWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload runtime_models.ExternalRegistrationStatusPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpAgentRegistrationClient) Self(ctx context.Context) (*runtime_models.ExternalRegistrationStatusPayload, error) {
	resp, err := c.client.PostInternalAgentRegistrationSelfWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload runtime_models.ExternalRegistrationStatusPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}

// mapRuntimeError maps the runtime's HTTP-status error to a domain error,
// preserving 400/404/409 and collapsing the rest to 502.
func mapRuntimeError(err error) error {
	var statusErr *helpers.RuntimeStatusError
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
