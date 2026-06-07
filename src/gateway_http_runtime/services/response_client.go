package services

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpResponseClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpResponseClient(client clients.ClientWithResponsesInterface) *HttpResponseClient {
	return &HttpResponseClient{client: client}
}

func (c *HttpResponseClient) Create(ctx context.Context, body *models.ExternalCreateResponsePayload) (*models.ExternalCreateResponseResultPayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalResponsesWithBodyWithResponse(ctx, jsonContentType, bytes.NewReader(encoded))
	if err != nil {
		return nil, err
	}
	var payload models.ExternalCreateResponseResultPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}
