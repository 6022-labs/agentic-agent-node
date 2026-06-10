package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpBlockchainConfigClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpBlockchainConfigClient(client clients.ClientWithResponsesInterface) *HttpBlockchainConfigClient {
	return &HttpBlockchainConfigClient{client: client}
}

func (c *HttpBlockchainConfigClient) Get(ctx context.Context) (*models.ExternalBlockchainConfigPayload, error) {
	resp, err := c.client.GetInternalConfigBlockchainWithResponse(ctx)
	if err != nil {
		return nil, mapRuntimeError(err)
	}
	var payload models.ExternalBlockchainConfigPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, mapRuntimeError(err)
	}
	return &payload, nil
}
