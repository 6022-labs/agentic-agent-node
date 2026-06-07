package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
)

type GetAgentMintInfo struct {
	client gateway_interfaces.AgentRegistrationClientInterface
}

func NewGetAgentMintInfo(client gateway_interfaces.AgentRegistrationClientInterface) *GetAgentMintInfo {
	return &GetAgentMintInfo{client: client}
}

func (u *GetAgentMintInfo) Execute(ctx context.Context) (*responses.MintInfoResponse, error) {
	payload, err := u.client.MintInfo(ctx)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *GetAgentMintInfo) toResponse(p *runtime_models.ExternalMintInfoPayload) *responses.MintInfoResponse {
	return &responses.MintInfoResponse{
		CollectionAddress: p.CollectionAddress,
		To:                p.To,
		Role:              p.Role,
		Name:              p.Name,
		Wallets:           p.Wallets,
		Images:            u.toKeyValues(p.Images),
		Attributes:        u.toKeyValues(p.Attributes),
		CloneOf:           p.CloneOf,
		ModeratorCount:    p.ModeratorCount,
		OwnerIsModerator:  p.OwnerIsModerator,
	}
}

func (u *GetAgentMintInfo) toKeyValues(in []runtime_models.ExternalKeyValue) []responses.KeyValue {
	out := make([]responses.KeyValue, 0, len(in))
	for _, kv := range in {
		out = append(out, responses.KeyValue{Key: kv.Key, Value: kv.Value})
	}
	return out
}
