package services

import (
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Collections scanned per findAgentByWallet call, to keep each RPC under the node limit.
const findAgentByWalletPageSize = 50

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

// FindAgentByWallet returns the agent NFT a wallet controls, or a zero address if none. The manager scans collections on-chain, paginated by the cursor.
func (r *AgentCollectionsManagerRequester) FindAgentByWallet(wallet common.Address) (common.Address, uint64, string, error) {
	contract, err := abi.NewAgentCollectionsManager(r.managersSettings.ManagerAddress, r.client)
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
