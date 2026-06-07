package use_cases

import (
	"context"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/use_cases/responses"
	runtime_models "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/models"
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

func (u *GetAgentEnsInfo) toResponse(p *runtime_models.ExternalEnsInfoPayload) *responses.EnsInfoResponse {
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
