package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpPaymentFacilitatorClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentFacilitatorClient(client clients.ClientWithResponsesInterface) *HttpPaymentFacilitatorClient {
	return &HttpPaymentFacilitatorClient{client: client}
}

func (c *HttpPaymentFacilitatorClient) Get(ctx context.Context) (map[string]models.ExternalPaymentFacilitator, error) {
	resp, err := c.client.GetInternalConfigPaymentFacilitatorsWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	var payload models.ExternalPaymentFacilitatorsPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	byNetwork := make(map[string]models.ExternalPaymentFacilitator, len(payload.Facilitators))
	for _, facilitator := range payload.Facilitators {
		byNetwork[facilitator.Network] = facilitator
	}
	return byNetwork, nil
}
