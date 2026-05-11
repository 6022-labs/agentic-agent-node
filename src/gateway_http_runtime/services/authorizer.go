package services

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
	"github.com/ethereum/go-ethereum/common"
)

const authorizePath = "/internal/authorize"

type HttpAuthorizer struct {
	client *clients.RuntimeHttpClient
}

func NewHttpAuthorizer(client *clients.RuntimeHttpClient) *HttpAuthorizer {
	return &HttpAuthorizer{client: client}
}

func (a *HttpAuthorizer) IsAuthorized(ctx context.Context, address common.Address) (bool, error) {
	body := runtime_models.ExternalAuthorizePayload{Address: address.Hex()}
	var payload runtime_models.ExternalAuthorizeResultPayload
	if err := a.client.Post(ctx, authorizePath, body, &payload); err != nil {
		return false, err
	}
	return payload.Authorized, nil
}
