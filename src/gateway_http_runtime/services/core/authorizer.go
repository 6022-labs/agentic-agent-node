package core

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/helpers"
	"github.com/ethereum/go-ethereum/common"
)

type HttpAuthorizer struct {
	client clients.ClientWithResponsesInterface
}

func NewHttpAuthorizer(client clients.ClientWithResponsesInterface) *HttpAuthorizer {
	return &HttpAuthorizer{client: client}
}

func (a *HttpAuthorizer) IsAuthorized(ctx context.Context, address common.Address) (bool, error) {
	body, err := json.Marshal(models.ExternalAuthorizePayload{Address: address.Hex()})
	if err != nil {
		return false, err
	}
	resp, err := a.client.PostInternalAuthorizeWithBodyWithResponse(ctx, helpers.JsonContentType, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	var payload models.ExternalAuthorizeResultPayload
	if err := helpers.DecodeRuntimeResponse(resp.HTTPResponse, resp.Body, &payload, nil); err != nil {
		return false, err
	}
	return payload.Authorized, nil
}
