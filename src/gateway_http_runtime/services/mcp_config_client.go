package services

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	"github.com/google/uuid"
)

const mcpConfigPath = "/internal/config/mcp"

type HttpMCPConfigClient struct {
	client *clients.RuntimeHttpClient
}

func NewHttpMCPConfigClient(client *clients.RuntimeHttpClient) *HttpMCPConfigClient {
	return &HttpMCPConfigClient{client: client}
}

func (c *HttpMCPConfigClient) List(ctx context.Context) (*runtime_models.MCPConnectionListPayload, error) {
	var payload runtime_models.MCPConnectionListPayload
	if err := c.client.Get(ctx, mcpConfigPath, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) Create(ctx context.Context, body *runtime_models.CreateMCPConnectionPayload) (*runtime_models.MCPConnectionPayload, error) {
	var payload runtime_models.MCPConnectionPayload
	if err := c.client.Post(ctx, mcpConfigPath, body, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) GetById(ctx context.Context, id uuid.UUID) (*runtime_models.MCPConnectionPayload, error) {
	var payload runtime_models.MCPConnectionPayload
	if err := c.client.Get(ctx, mcpConfigPath+"/"+id.String(), &payload, gateway_interfaces.ErrMCPConnectionNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) Update(ctx context.Context, id uuid.UUID, body *runtime_models.UpdateMCPConnectionPayload) (*runtime_models.MCPConnectionPayload, error) {
	var payload runtime_models.MCPConnectionPayload
	if err := c.client.Put(ctx, mcpConfigPath+"/"+id.String(), body, &payload, gateway_interfaces.ErrMCPConnectionNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) Delete(ctx context.Context, id uuid.UUID) error {
	return c.client.Delete(ctx, mcpConfigPath+"/"+id.String(), nil)
}
