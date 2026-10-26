package payment

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpPaymentReceivePolicyClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentReceivePolicyClient(client clients.ClientWithResponsesInterface) *HttpPaymentReceivePolicyClient {
	return &HttpPaymentReceivePolicyClient{client: client}
}

func (c *HttpPaymentReceivePolicyClient) Get(ctx context.Context) (*models.ExternalPaymentReceivePolicyPayload, error) {
	resp, err := c.client.GetInternalConfigPaymentReceivePolicyWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentReceivePolicyPayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpPaymentReceivePolicyClient) Delete(ctx context.Context) error {
	resp, err := c.client.DeleteInternalConfigPaymentReceivePolicyWithResponse(ctx)
	if err != nil {
		return helpers.MapRuntimeError(err)
	}
	return helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil)
}
