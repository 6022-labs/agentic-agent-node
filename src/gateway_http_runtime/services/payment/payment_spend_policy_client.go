package payment

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpPaymentSpendPolicyClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentSpendPolicyClient(client clients.ClientWithResponsesInterface) *HttpPaymentSpendPolicyClient {
	return &HttpPaymentSpendPolicyClient{client: client}
}

func (c *HttpPaymentSpendPolicyClient) Get(ctx context.Context) (*models.ExternalPaymentSpendPolicyPayload, error) {
	resp, err := c.client.GetInternalConfigPaymentSpendPolicyWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentSpendPolicyPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &payload, nil
}

func (c *HttpPaymentSpendPolicyClient) Update(
	ctx context.Context,
	payload *models.ExternalUpdatePaymentSpendPolicyPayload,
) (*models.ExternalPaymentSpendPolicyPayload, error) {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalConfigPaymentSpendPolicyWithBodyWithResponse(
		ctx,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var policy models.ExternalPaymentSpendPolicyPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &policy, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &policy, nil
}

func (c *HttpPaymentSpendPolicyClient) Delete(ctx context.Context) error {
	resp, err := c.client.DeleteInternalConfigPaymentSpendPolicyWithResponse(ctx)
	if err != nil {
		return helpers.MapRuntimeError(err)
	}
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil); err != nil {
		return helpers.MapRuntimeError(err)
	}
	return nil
}
