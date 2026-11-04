package payment

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
)

type HttpPaymentSpendRuleClient struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpPaymentSpendRuleClient(client clients.ClientWithResponsesInterface) *HttpPaymentSpendRuleClient {
	return &HttpPaymentSpendRuleClient{client: client}
}

func (c *HttpPaymentSpendRuleClient) List(ctx context.Context) ([]models.ExternalPaymentSpendRule, error) {
	resp, err := c.client.GetInternalConfigPaymentSpendPolicyRulesWithResponse(ctx)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalListPaymentSpendRulesPayload
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return payload.Rules, nil
}

func (c *HttpPaymentSpendRuleClient) Create(
	ctx context.Context,
	rule *models.ExternalPaymentSpendRule,
) (*models.ExternalPaymentSpendRule, error) {
	encoded, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PostInternalConfigPaymentSpendPolicyRulesWithBodyWithResponse(
		ctx,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentSpendRule
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpPaymentSpendRuleClient) Update(
	ctx context.Context,
	id string,
	rule *models.ExternalPaymentSpendRule,
) (*models.ExternalPaymentSpendRule, error) {
	encoded, err := json.Marshal(rule)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.PutInternalConfigPaymentSpendPolicyRulesIdWithBodyWithResponse(
		ctx,
		id,
		helpers.JsonContentType,
		bytes.NewReader(encoded),
	)
	if err != nil {
		return nil, helpers.MapRuntimeError(err)
	}
	var payload models.ExternalPaymentSpendRule
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return nil, err
	}
	return &payload, nil
}

func (c *HttpPaymentSpendRuleClient) Delete(ctx context.Context, id string) error {
	resp, err := c.client.DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(ctx, id)
	if err != nil {
		return helpers.MapRuntimeError(err)
	}
	if err := helpers.DecodeAndMapRuntimeResponse(resp.HTTPResponse, resp.Body, nil, nil); err != nil {
		return err
	}
	return nil
}
