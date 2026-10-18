package identity

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpAgentCollectionsClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAgentCollectionsClient(client clients.ClientWithResponsesInterface) *HttpAgentCollectionsClient {
	return &HttpAgentCollectionsClient{client: client}
}

func (c *HttpAgentCollectionsClient) List(
	ctx context.Context,
	chainId string,
) (*models.ExternalAgentCollectionsPayload, error) {
	resp, err := c.client.GetInternalAgentCollectionsWithResponse(ctx, &clients.GetInternalAgentCollectionsParams{
		ChainId: chainId,
	})
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAgentCollectionsPayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}
