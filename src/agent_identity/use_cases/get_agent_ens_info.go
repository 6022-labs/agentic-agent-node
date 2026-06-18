package use_cases

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases/responses"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
)

// GetAgentEnsInfo assembles the initAgentTexts calldata so the connected wallet can
// register the agent's ENS records itself. Requires the NFT to be minted.
type GetAgentEnsInfo struct {
	agentRepository        repositories.AgentRepositoryInterface
	runtimeRecordsProvider identity_interfaces.RuntimeRecordsProviderInterface
}

func NewGetAgentEnsInfo(
	agentRepository repositories.AgentRepositoryInterface,
	runtimeRecordsProvider identity_interfaces.RuntimeRecordsProviderInterface,
) *GetAgentEnsInfo {
	return &GetAgentEnsInfo{agentRepository: agentRepository, runtimeRecordsProvider: runtimeRecordsProvider}
}

func (u *GetAgentEnsInfo) Execute(ctx context.Context) (*responses.EnsInfoResponse, error) {
	agent, err := u.agentRepository.Get(ctx)
	if err != nil {
		return nil, errors.NewInternalError("agent_lookup_failed", err.Error())
	}
	if agent == nil {
		return nil, errors.NewNotFoundError("draft_not_found", "no agent registration has been saved yet")
	}
	if agent.CollectionTokenId == nil {
		return nil, errors.NewConflictError("not_minted", "agent NFT must be minted before ENS registration")
	}

	records := make([]responses.KeyValue, 0)
	for _, rec := range u.runtimeRecordsProvider.RuntimeRecords() {
		records = append(records, responses.KeyValue{Key: rec.Key, Value: rec.Value})
	}

	return &responses.EnsInfoResponse{
		CollectionAddress: common.HexToAddress(agent.CollectionAddress).Hex(),
		TokenId:           *agent.CollectionTokenId,
		Records:           records,
	}, nil
}
