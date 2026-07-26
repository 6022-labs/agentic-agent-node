package reasoning

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpLLMConfigClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpLLMConfigClient(client clients.ClientWithResponsesInterface) *HttpLLMConfigClient {
	return &HttpLLMConfigClient{client: client}
}

func (c *HttpLLMConfigClient) Get(ctx context.Context) (*models.ExternalLLMConnectionPayload, error) {
	resp, err := c.client.GetInternalConfigLlmWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalLLMConnectionPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, gateway_interfaces.ErrLLMConnectionNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpLLMConfigClient) Setup(ctx context.Context, body *models.ExternalSetupLLMConnectionPayload) (*models.ExternalLLMConnectionPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConfigLlmWithBodyWithResponse(ctx, helpers.JsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalLLMConnectionPayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpLLMConfigClient) Validate(ctx context.Context, body *models.ExternalSetupLLMConnectionPayload) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := c.client.PostInternalConfigLlmValidateWithBodyWithResponse(ctx, helpers.JsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return helpers.MapRuntimeError(err)
	}
	return helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil)
}

func (c *HttpLLMConfigClient) Delete(ctx context.Context) error {
	resp, err := c.client.DeleteInternalConfigLlmWithResponse(ctx)
	if err != nil {
		return err
	}
	return helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil)
}
