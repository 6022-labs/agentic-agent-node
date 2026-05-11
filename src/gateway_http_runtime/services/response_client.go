package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

const responsesPath = "/internal/responses"

type HttpResponseClient struct {
	client *clients.RuntimeHttpClient
}

func NewHttpResponseClient(client *clients.RuntimeHttpClient) *HttpResponseClient {
	return &HttpResponseClient{client: client}
}

func (c *HttpResponseClient) Create(ctx context.Context, body *runtime_models.ExternalCreateResponsePayload) (*runtime_models.ExternalCreateResponseResultPayload, error) {
	var payload runtime_models.ExternalCreateResponseResultPayload
	if err := c.client.Post(ctx, responsesPath, body, &payload); err != nil {
		return nil, err
	}
	return &payload, nil
}
