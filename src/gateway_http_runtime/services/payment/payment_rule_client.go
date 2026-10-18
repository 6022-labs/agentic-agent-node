package payment

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpPaymentRuleClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentRuleClient(client clients.ClientWithResponsesInterface) *HttpPaymentRuleClient {
	return &HttpPaymentRuleClient{client: client}
}

func (c *HttpPaymentRuleClient) List(ctx context.Context) ([]models.ExternalPaymentRule, error) {
	resp, err := c.client.GetInternalConfigPaymentPolicyRulesWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalListPaymentRulesPayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return payload.Rules, nil
}

func (c *HttpPaymentRuleClient) Create(
	ctx context.Context,
	rule *models.ExternalPaymentRule,
) (*models.ExternalPaymentRule, error) {
	encoded, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConfigPaymentPolicyRulesWithBodyWithResponse(
		ctx,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentRule
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpPaymentRuleClient) Update(
	ctx context.Context,
	id string,
	rule *models.ExternalPaymentRule,
) (*models.ExternalPaymentRule, error) {
	encoded, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalConfigPaymentPolicyRulesIdWithBodyWithResponse(
		ctx,
		id,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentRule
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpPaymentRuleClient) Delete(ctx context.Context, id string) error {
	resp, err := c.client.DeleteInternalConfigPaymentPolicyRulesIdWithResponse(ctx, id)
	if err != nil {
		return helpers.MapRuntimeError(err)
	}
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil); err != nil {
		return err
	}
	return nil
}
