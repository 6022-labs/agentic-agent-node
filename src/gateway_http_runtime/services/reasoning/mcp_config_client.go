package reasoning

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/google/uuid"
)

type HttpMCPConfigClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpMCPConfigClient(client clients.ClientWithResponsesInterface) *HttpMCPConfigClient {
	return &HttpMCPConfigClient{client: client}
}

func (c *HttpMCPConfigClient) List(ctx context.Context) (*models.ExternalMCPConnectionListPayload, error) {
	resp, err := c.client.GetInternalConfigMcpWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalMCPConnectionListPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) Create(ctx context.Context, body *models.ExternalCreateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConfigMcpWithBodyWithResponse(ctx, helpers.JsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	var payload models.ExternalMCPConnectionPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) GetById(ctx context.Context, id uuid.UUID) (*models.ExternalMCPConnectionPayload, error) {
	resp, err := c.client.GetInternalConfigMcpIdWithResponse(ctx, id.String())
	if err != nil {
		return nil, err
	}
	var payload models.ExternalMCPConnectionPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, gateway_interfaces.ErrMCPConnectionNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) Update(ctx context.Context, id uuid.UUID, body *models.ExternalUpdateMCPConnectionPayload) (*models.ExternalMCPConnectionPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalConfigMcpIdWithBodyWithResponse(ctx, id.String(), helpers.JsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	var payload models.ExternalMCPConnectionPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, gateway_interfaces.ErrMCPConnectionNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpMCPConfigClient) Delete(ctx context.Context, id uuid.UUID) error {
	resp, err := c.client.DeleteInternalConfigMcpIdWithResponse(ctx, id.String())
	if err != nil {
		return err
	}
	return helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil)
}
