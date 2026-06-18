// Package factory builds blockchain primitives shared by the on-chain writer
// adapters (mint, ENS).
package factory

import (
	"context"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_blockchain/helper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// TransactorFactory produces chain-bound transactors signed by the node's Vault
// wallet, shared by the on-chain writer adapters so the wallet-load +
// transactor-build sequence lives in one place.
type TransactorFactory struct {
	walletStorage   identity_interfaces.WalletCredentialsStorageInterface
	agentRepository repositories.AgentRepositoryInterface
}

func NewTransactorFactory(
	walletStorage identity_interfaces.WalletCredentialsStorageInterface,
	agentRepository repositories.AgentRepositoryInterface,
) *TransactorFactory {
	return &TransactorFactory{
		walletStorage:   walletStorage,
		agentRepository: agentRepository,
	}
}

// Create loads the node wallet from Vault and builds a transactor bound to the
// given chain, scoped to ctx.
func (f *TransactorFactory) Create(ctx context.Context, chainId uint64) (*bind.TransactOpts, error) {
	agent, err := f.agentRepository.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("load agent: %w", err)
	}
	if agent == nil {
		return nil, fmt.Errorf("agent not bootstrapped")
	}
	cred, found, err := f.walletStorage.GetForId(ctx, agent.Id.String())
	if err != nil {
		return nil, fmt.Errorf("load wallet credentials: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("wallet credentials not present; bootstrap must complete before signing")
	}

	opts, err := helper.Transactor(cred.PrivateKeyHex, chainId)
	if err != nil {
		return nil, err
	}
	opts.Context = ctx
	return opts, nil
}
