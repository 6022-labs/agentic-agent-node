package services

import (
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type AgentCollectionRequester struct {
	client *ethclient.Client
}

func NewAgentCollectionRequester(client *ethclient.Client) *AgentCollectionRequester {
	return &AgentCollectionRequester{client: client}
}

func (r *AgentCollectionRequester) OwnerOf(collectionAddress common.Address, tokenId uint64) (common.Address, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return common.Address{}, err
	}
	return contract.OwnerOf(nil, new(big.Int).SetUint64(tokenId))
}

func (r *AgentCollectionRequester) CollectionName(collectionAddress common.Address) (string, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return "", err
	}
	return contract.Name(nil)
}
