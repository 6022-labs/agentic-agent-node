package reasoning

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpAgentGuidanceClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAgentGuidanceClient(client clients.ClientWithResponsesInterface) *HttpAgentGuidanceClient {
	return &HttpAgentGuidanceClient{client: client}
}

func (c *HttpAgentGuidanceClient) Get(ctx context.Context) (*models.ExternalAgentGuidancePayload, error) {
	resp, err := c.client.GetInternalAgentGuidanceWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAgentGuidancePayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpAgentGuidanceClient) Set(
	ctx context.Context,
	body *models.ExternalAgentGuidancePayload,
) (*models.ExternalAgentGuidancePayload, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalAgentGuidanceWithBodyWithResponse(
		ctx,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAgentGuidancePayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}
