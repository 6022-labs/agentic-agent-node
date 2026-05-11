package services

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

const llmConfigPath = "/internal/config/llm"

type HttpLLMConfigClient struct {
	client *clients.RuntimeHttpClient
}

func NewHttpLLMConfigClient(client *clients.RuntimeHttpClient) *HttpLLMConfigClient {
	return &HttpLLMConfigClient{client: client}
}

func (c *HttpLLMConfigClient) Get(ctx context.Context) (*runtime_models.ExternalLLMConnectionPayload, error) {
	var payload runtime_models.ExternalLLMConnectionPayload
	if err := c.client.Get(ctx, llmConfigPath, &payload, gateway_interfaces.ErrLLMConnectionNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpLLMConfigClient) Setup(ctx context.Context, body *runtime_models.ExternalSetupLLMConnectionPayload) (*runtime_models.ExternalLLMConnectionPayload, error) {
	var payload runtime_models.ExternalLLMConnectionPayload
	if err := c.client.Post(ctx, llmConfigPath, body, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpLLMConfigClient) Delete(ctx context.Context) error {
	return c.client.Delete(ctx, llmConfigPath, nil)
}
