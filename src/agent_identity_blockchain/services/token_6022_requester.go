package services

import (
	"fmt"
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/factory"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
)

type Token6022Requester struct {
	ethClientFactory  *factory.EthClientFactory
	token6022Settings *settings.Token6022Settings
}

func NewToken6022Requester(
	ethClientFactory *factory.EthClientFactory,
	token6022Settings *settings.Token6022Settings,
) *Token6022Requester {
	return &Token6022Requester{
		ethClientFactory:  ethClientFactory,
		token6022Settings: token6022Settings,
	}
}

func (r *Token6022Requester) token(chainId string) (*abi.Token6022, error) {
	address, ok := r.token6022Settings.AddressFor(chainId)
	if !ok {
		return nil, fmt.Errorf("no $6022 token configured for chainId %s", chainId)
	}
	client, err := r.ethClientFactory.For(chainId)
	if err != nil {
		return nil, err
	}
	return abi.NewToken6022(address, client)
}

func (r *Token6022Requester) BalanceOf(chainId string, account common.Address) (*big.Int, error) {
	token6022, err := r.token(chainId)
	if err != nil {
		return nil, err
	}
	return token6022.BalanceOf(nil, account)
}

// token6022Decimals is fixed at 18, so we skip the on-chain decimals() call.
const token6022Decimals = 18

func (r *Token6022Requester) Decimals() (uint8, error) {
	return token6022Decimals, nil
}
