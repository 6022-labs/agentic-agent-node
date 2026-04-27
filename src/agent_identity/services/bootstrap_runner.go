package services

import (
	"context"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"go.uber.org/zap"
)

// BootstrapRunner ensures the agent's wallet is loaded from (or generated
// into) the local Vault. Once that single step succeeds the agent is "ready"
// from a wallet-state point of view; on-chain identity is resolved lazily by
// AgentIdentityResolver as routes need it.
type BootstrapRunner struct {
	logger        *zap.Logger
	walletStorage identity_interfaces.WalletCredentialsStorageInterface
	stateHolder   identity_interfaces.BootstrapStateHolderInterface
}

func NewBootstrapRunner(
	logger *zap.Logger,
	walletStorage identity_interfaces.WalletCredentialsStorageInterface,
	stateHolder identity_interfaces.BootstrapStateHolderInterface,
) *BootstrapRunner {
	return &BootstrapRunner{
		logger:        logger,
		walletStorage: walletStorage,
		stateHolder:   stateHolder,
	}
}

// Run loads (or generates) the wallet and moves the holder to ready.
func (r *BootstrapRunner) Run(ctx context.Context) error {
	cred, err := r.walletStorage.GetOrGenerate(ctx)
	if err != nil {
		msg := fmt.Sprintf("wallet bootstrap failed: %v", err)
		r.logger.Error(msg)
		r.stateHolder.SetError(msg)
		return err
	}

	r.stateHolder.SetReady(cred.Address)
	r.logger.Info("wallet ready", zap.String("address", cred.Address))
	return nil
}
