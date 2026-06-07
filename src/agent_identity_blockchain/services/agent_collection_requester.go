package services

import (
	"math/big"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
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

func (r *AgentCollectionRequester) WalletToTokenId(collectionAddress common.Address, wallet common.Address) (uint64, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return 0, err
	}
	id, err := contract.WalletToTokenId(nil, wallet)
	if err != nil {
		return 0, err
	}
	return id.Uint64(), nil
}

func (r *AgentCollectionRequester) NameToTokenId(collectionAddress common.Address, name string) (uint64, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return 0, err
	}
	id, err := contract.NameToTokenId(nil, name)
	if err != nil {
		return 0, err
	}
	return id.Uint64(), nil
}

func (r *AgentCollectionRequester) InformationOf(collectionAddress common.Address, tokenId uint64) (identity_interfaces.OnChainAgentInfo, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return identity_interfaces.OnChainAgentInfo{}, err
	}
	info, err := contract.InformationOf(nil, new(big.Int).SetUint64(tokenId))
	if err != nil {
		return identity_interfaces.OnChainAgentInfo{}, err
	}

	var cloneOf *uint64
	if info.CloneOf.IsSet && info.CloneOf.Value != nil {
		v := info.CloneOf.Value.Uint64()
		cloneOf = &v
	}
	return identity_interfaces.OnChainAgentInfo{
		Name:    info.Name,
		Role:    info.Role,
		Creator: info.Creator,
		CloneOf: cloneOf,
	}, nil
}

func (r *AgentCollectionRequester) ModeratorCount(collectionAddress common.Address) (uint64, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return 0, err
	}
	count, err := contract.ModeratorCount(nil)
	if err != nil {
		return 0, err
	}
	return count.Uint64(), nil
}

func (r *AgentCollectionRequester) IsModerator(collectionAddress common.Address, account common.Address) (bool, error) {
	contract, err := abi.NewAgentCollectionV1(collectionAddress, r.client)
	if err != nil {
		return false, err
	}
	return contract.IsModerator(nil, account)
}
