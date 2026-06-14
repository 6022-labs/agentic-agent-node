package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/models"
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/identity/use_cases/responses"
)

type GetAgentEnsInfo struct {
	client gateway_interfaces.AgentRegistrationClientInterface
}

func NewGetAgentEnsInfo(client gateway_interfaces.AgentRegistrationClientInterface) *GetAgentEnsInfo {
	return &GetAgentEnsInfo{client: client}
}

func (u *GetAgentEnsInfo) Execute(ctx context.Context) (*responses.EnsInfoResponse, error) {
	payload, err := u.client.EnsInfo(ctx)
	if err != nil {
		return nil, err
	}
	return u.toResponse(payload), nil
}

func (u *GetAgentEnsInfo) toResponse(p *models.ExternalEnsInfoPayload) *responses.EnsInfoResponse {
	records := make([]responses.KeyValue, 0, len(p.Records))
	for _, kv := range p.Records {
		records = append(records, responses.KeyValue{Key: kv.Key, Value: kv.Value})
	}
	return &responses.EnsInfoResponse{
		CollectionAddress: p.CollectionAddress,
		TokenId:           p.TokenId,
		Records:           records,
	}
}
