package services

import (
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type AgentCollectionsManagerRequester struct {
	client           *ethclient.Client
	managersSettings *settings.AgentCollectionsManagersSettings
}

func NewAgentCollectionsManagerRequester(
	client *ethclient.Client,
	managersSettings *settings.AgentCollectionsManagersSettings,
) *AgentCollectionsManagerRequester {
	return &AgentCollectionsManagerRequester{
		client:           client,
		managersSettings: managersSettings,
	}
}

func (r *AgentCollectionsManagerRequester) ListAllCollections() ([]common.Address, error) {
	contract, err := abi.NewAgentCollectionsManager(r.managersSettings.ManagerAddress, r.client)
	if err != nil {
		return nil, err
	}

	nextIdBig, err := contract.NextCollectionId(nil)
	if err != nil {
		return nil, err
	}
	if nextIdBig == nil || !nextIdBig.IsUint64() {
		return nil, nil
	}
	nextId := nextIdBig.Uint64()

	out := make([]common.Address, 0, nextId)
	for i := uint64(0); i < nextId; i++ {
		addr, err := contract.Collections(nil, new(big.Int).SetUint64(i))
		if err != nil {
			return nil, err
		}
		if (addr == common.Address{}) {
			continue
		}
		out = append(out, addr)
	}
	return out, nil
}
