package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpAgentCollectionsClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAgentCollectionsClient(client clients.ClientWithResponsesInterface) *HttpAgentCollectionsClient {
	return &HttpAgentCollectionsClient{client: client}
}

func (c *HttpAgentCollectionsClient) List(ctx context.Context) (*models.ExternalAgentCollectionsPayload, error) {
	resp, err := c.client.GetInternalAgentCollectionsWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload models.ExternalAgentCollectionsPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}
