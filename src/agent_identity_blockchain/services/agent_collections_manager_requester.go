package services

import (
	"fmt"
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
)

// Collections scanned per findAgentByWallet call, to keep each RPC under the node limit.
const findAgentByWalletPageSize = 50

// Collections fetched per listCollections call, to keep each RPC under the node limit.
const listCollectionsPageSize = 50

type AgentCollectionsManagerRequester struct {
	ethClientFactory *factory.EthClientFactory
	managersSettings *settings.AgentCollectionsManagersSettings
}

func NewAgentCollectionsManagerRequester(
	ethClientFactory *factory.EthClientFactory,
	managersSettings *settings.AgentCollectionsManagersSettings,
) *AgentCollectionsManagerRequester {
	return &AgentCollectionsManagerRequester{
		ethClientFactory: ethClientFactory,
		managersSettings: managersSettings,
	}
}

func (r *AgentCollectionsManagerRequester) manager(chainId string) (*abi.AgentCollectionsManager, error) {
	address, ok := r.managersSettings.AddressFor(chainId)
	if !ok {
		return nil, fmt.Errorf("no AgentCollectionsManager configured for chainId %s", chainId)
	}
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return nil, err
	}
	return abi.NewAgentCollectionsManager(address, client)
}

// IsKnownCollection reports whether the manager governs this collection.
func (r *AgentCollectionsManagerRequester) IsKnownCollection(chainId string, collection common.Address) (bool, error) {
	contract, err := r.manager(chainId)
	if err != nil {
		return false, err
	}
	return contract.IsKnownCollection(nil, collection)
}

// FindAgentByWallet returns the agent NFT a wallet controls, or a zero address if none. The manager scans collections on-chain, paginated by the cursor.
func (r *AgentCollectionsManagerRequester) FindAgentByWallet(chainId string, wallet common.Address) (common.Address, uint64, string, error) {
	contract, err := r.manager(chainId)
	if err != nil {
		return common.Address{}, 0, "", err
	}

	limit := new(big.Int).SetUint64(findAgentByWalletPageSize)
	for offset := big.NewInt(0); ; {
		res, err := contract.FindAgentByWallet(nil, wallet, offset, limit)
		if err != nil {
			return common.Address{}, 0, "", err
		}

		if (res.Collection != common.Address{}) {
			tokenId := uint64(0)
			if res.TokenId != nil && res.TokenId.IsUint64() {
				tokenId = res.TokenId.Uint64()
			}
			return res.Collection, tokenId, res.Name, nil
		}

		// next == 0 means the scan reached the last collection without a match.
		if res.Next == nil || res.Next.Sign() == 0 {
			return common.Address{}, 0, "", nil
		}
		offset = res.Next
	}
}

func (r *AgentCollectionsManagerRequester) ListCollections(chainId string) ([]models.AgentCollection, error) {
	contract, err := r.manager(chainId)
	if err != nil {
		return nil, err
	}

	limit := new(big.Int).SetUint64(listCollectionsPageSize)
	collections := make([]models.AgentCollection, 0)
	for offset := big.NewInt(0); ; {
		res, err := contract.ListCollections(nil, offset, limit)
		if err != nil {
			return nil, err
		}

		for _, info := range res.Infos {
			collections = append(collections, models.AgentCollection{
				Address:        info.CollectionAddress.Hex(),
				Name:           info.Name,
				Description:    info.Description,
				ModeratorCount: helper.ToUint64(info.ModeratorCount),
				NextTokenId:    helper.ToUint64(info.NextTokenId),
			})
		}

		// next == 0 means the listing reached the last collection.
		if res.Next == nil || res.Next.Sign() == 0 {
			return collections, nil
		}
		offset = res.Next
	}
}
