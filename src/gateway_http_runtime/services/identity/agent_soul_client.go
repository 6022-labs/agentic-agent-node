package identity

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpAgentSoulClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAgentSoulClient(client clients.ClientWithResponsesInterface) *HttpAgentSoulClient {
	return &HttpAgentSoulClient{client: client}
}

func (c *HttpAgentSoulClient) Get(ctx context.Context) (*models.ExternalAgentSoulPayload, error) {
	resp, err := c.client.GetInternalAgentSoulWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAgentSoulPayload
	if err := helpers.DecodeAndMapRuntimeResponse(
		resp.HTTPResponse,
		resp.Body,
		&payload,
		identity_interfaces.ErrAgentSoulNotFound,
	); err != nil {
		return nil, err
	}
	return &payload, nil
}
