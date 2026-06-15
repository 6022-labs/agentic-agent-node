package factory

import (
	"context"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity_ens/helper"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

// TransactorFactory builds chain-bound transactors signed by the node's Vault
// wallet, for this adapter's writes.
type TransactorFactory struct {
	walletStorage identity_interfaces.WalletCredentialsStorageInterface
}

func NewTransactorFactory(walletStorage identity_interfaces.WalletCredentialsStorageInterface) *TransactorFactory {
	return &TransactorFactory{walletStorage: walletStorage}
}

// Create loads the node wallet and builds a transactor bound to the chain.
func (f *TransactorFactory) Create(ctx context.Context, chainId uint64) (*bind.TransactOpts, error) {
	cred, found, err := f.walletStorage.Get(ctx)
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
