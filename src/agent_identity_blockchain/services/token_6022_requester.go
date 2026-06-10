package services

import (
	"math/big"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/abi"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Token6022Requester struct {
	client            *ethclient.Client
	token6022Settings *settings.Token6022Settings
}

func NewToken6022Requester(client *ethclient.Client, token6022Settings *settings.Token6022Settings) *Token6022Requester {
	return &Token6022Requester{
		client:            client,
		token6022Settings: token6022Settings,
	}
}

func (r *Token6022Requester) BalanceOf(account common.Address) (*big.Int, error) {
	token6022, err := abi.NewToken6022(r.token6022Settings.SmartContractAddress, r.client)
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
