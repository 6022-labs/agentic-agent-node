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

func (r *AgentCollectionRequester) TokenIdByWallet(collectionAddress common.Address, wallet common.Address) (uint64, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return 0, err
	}
	tokenId, err := contract.WalletToTokenId(nil, wallet)
	if err != nil {
		return 0, err
	}
	if tokenId == nil || !tokenId.IsUint64() {
		return 0, nil
	}
	return tokenId.Uint64(), nil
}

func (r *AgentCollectionRequester) OwnerOf(collectionAddress common.Address, tokenId uint64) (common.Address, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return common.Address{}, err
	}
	return contract.OwnerOf(nil, new(big.Int).SetUint64(tokenId))
}

func (r *AgentCollectionRequester) NameOf(collectionAddress common.Address, tokenId uint64) (string, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return "", err
	}
	return contract.NameOf(nil, new(big.Int).SetUint64(tokenId))
}

func (r *AgentCollectionRequester) CollectionName(collectionAddress common.Address) (string, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return "", err
	}
	return contract.Name(nil)
}
