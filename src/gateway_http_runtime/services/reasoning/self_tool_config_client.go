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

type HttpSelfToolConfigClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpSelfToolConfigClient(client clients.ClientWithResponsesInterface) *HttpSelfToolConfigClient {
	return &HttpSelfToolConfigClient{client: client}
}

func (c *HttpSelfToolConfigClient) List(ctx context.Context) (*models.ExternalSelfToolListPayload, error) {
	resp, err := c.client.GetInternalConfigSelfToolsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalSelfToolListPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpSelfToolConfigClient) Update(
	ctx context.Context,
	id uuid.UUID,
	body *models.ExternalUpdateSelfToolPayload,
) (*models.ExternalSelfToolPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalConfigSelfToolsIdWithBodyWithResponse(
		ctx,
		id.String(),
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalSelfToolPayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, gateway_interfaces.ErrSelfToolNotFound); err != nil {
		return nil, err
	}
	return &payload, nil
}
