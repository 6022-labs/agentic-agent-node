package payment

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type authorizePaymentReceiveAccessBody struct {
	Address          string `json:"address,omitempty"`
	PaymentSignature string `json:"paymentSignature,omitempty"`
}

type HttpPaymentReceiveAccessClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentReceiveAccessClient(client clients.ClientWithResponsesInterface) *HttpPaymentReceiveAccessClient {
	return &HttpPaymentReceiveAccessClient{client: client}
}

func (c *HttpPaymentReceiveAccessClient) Authorize(
	ctx context.Context,
	address string,
	paymentSignature string,
) (*models.ExternalPaymentReceiveAccessDecision, error) {
	// Hand-marshalled: the generated body is a union type.
	body, err := json.Marshal(authorizePaymentReceiveAccessBody{Address: address, PaymentSignature: paymentSignature})
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}

	resp, err := c.client.PostInternalPaymentReceiveAccessWithBodyWithResponse(
		ctx,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var decision models.ExternalPaymentReceiveAccessDecision
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &decision, nil); err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	return &decision, nil
}
