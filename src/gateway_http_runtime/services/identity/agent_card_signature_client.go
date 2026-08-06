package identity

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type signAgentCardBody struct {
	Card json.RawMessage `json:"card"`
}

type HttpAgentCardSignatureClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAgentCardSignatureClient(client clients.ClientWithResponsesInterface) *HttpAgentCardSignatureClient {
	return &HttpAgentCardSignatureClient{client: client}
}

func (c *HttpAgentCardSignatureClient) Sign(ctx context.Context, card []byte) (*models.ExternalAgentCardSignaturePayload, error) {
	// Raw JSON, not the generated map body, so the runtime signs exactly these bytes.
	body, err := json.Marshal(signAgentCardBody{Card: card})
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}

	resp, err := c.client.PostInternalIdentityCardSignatureWithBodyWithResponse(ctx, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalAgentCardSignaturePayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}
