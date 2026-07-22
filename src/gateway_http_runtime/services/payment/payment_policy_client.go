package payment

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpPaymentPolicyClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentPolicyClient(client clients.ClientWithResponsesInterface) *HttpPaymentPolicyClient {
	return &HttpPaymentPolicyClient{client: client}
}

func (c *HttpPaymentPolicyClient) Get(ctx context.Context) (*models.ExternalPaymentPolicyPayload, error) {
	resp, err := c.client.GetInternalConfigPaymentPolicyWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentPolicyPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpPaymentPolicyClient) Delete(ctx context.Context) error {
	resp, err := c.client.DeleteInternalConfigPaymentPolicyWithResponse(ctx)
	if err != nil {
		return err
	}
	return helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil)
}
