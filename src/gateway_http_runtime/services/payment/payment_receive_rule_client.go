package payment

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpPaymentReceiveRuleClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentReceiveRuleClient(client clients.ClientWithResponsesInterface) *HttpPaymentReceiveRuleClient {
	return &HttpPaymentReceiveRuleClient{client: client}
}

func (c *HttpPaymentReceiveRuleClient) List(ctx context.Context) ([]models.ExternalPaymentReceiveRule, error) {
	resp, err := c.client.GetInternalConfigPaymentReceivePolicyRulesWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalListPaymentReceiveRulesPayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return payload.Rules, nil
}

func (c *HttpPaymentReceiveRuleClient) Create(
	ctx context.Context,
	rule *models.ExternalPaymentReceiveRule,
) (*models.ExternalPaymentReceiveRule, error) {
	encoded, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConfigPaymentReceivePolicyRulesWithBodyWithResponse(
		ctx,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentReceiveRule
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpPaymentReceiveRuleClient) Update(
	ctx context.Context,
	id string,
	rule *models.ExternalPaymentReceiveRule,
) (*models.ExternalPaymentReceiveRule, error) {
	encoded, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalConfigPaymentReceivePolicyRulesIdWithBodyWithResponse(
		ctx,
		id,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentReceiveRule
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpPaymentReceiveRuleClient) Delete(ctx context.Context, id string) error {
	resp, err := c.client.DeleteInternalConfigPaymentReceivePolicyRulesIdWithResponse(ctx, id)
	if err != nil {
		return helpers.MapRuntimeError(err)
	}
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil); err != nil {
		return err
	}
	return nil
}
