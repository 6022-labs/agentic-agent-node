package services

import (
	"bytes"
	"context"
	"encoding/json"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type HttpLLMConfigClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpLLMConfigClient(client clients.ClientWithResponsesInterface) *HttpLLMConfigClient {
	return &HttpLLMConfigClient{client: client}
}

func (c *HttpLLMConfigClient) Get(ctx context.Context) (*runtime_models.ExternalLLMConnectionPayload, error) {
	resp, err := c.client.GetInternalConfigLlmWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	var payload runtime_models.ExternalLLMConnectionPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, gateway_interfaces.ErrLLMConnectionNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpLLMConfigClient) Setup(ctx context.Context, body *runtime_models.ExternalSetupLLMConnectionPayload) (*runtime_models.ExternalLLMConnectionPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConfigLlmWithBodyWithResponse(ctx, jsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	var payload runtime_models.ExternalLLMConnectionPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpLLMConfigClient) Delete(ctx context.Context) error {
	resp, err := c.client.DeleteInternalConfigLlmWithResponse(ctx)
	if err != nil {
		return err
	}
	return helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil)
}
