package services

import (
	"context"

	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/common/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

const identityPath = "/internal/identity"

type HttpAgentIdentityResolver struct {
	client *clients.RuntimeHttpClient
}

func NewHttpAgentIdentityResolver(client *clients.RuntimeHttpClient) *HttpAgentIdentityResolver {
	return &HttpAgentIdentityResolver{client: client}
}

func (r *HttpAgentIdentityResolver) Resolve(ctx context.Context) (*models.AgentIdentity, error) {
	var payload runtime_models.AgentIdentityPayload
	if err := r.client.Get(ctx, identityPath, &payload, common_interfaces.ErrIdentityNotFound); err != nil {
		return nil, err
	}
	return &models.AgentIdentity{
		CollectionAddress: payload.CollectionAddress,
		TokenId:           payload.TokenId,
		Name:              payload.Name,
	}, nil
}
