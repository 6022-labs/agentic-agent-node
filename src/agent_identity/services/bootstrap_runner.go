package services

import (
	"context"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"go.uber.org/zap"
)

// BootstrapRunner loads or generates the agent's wallet in the local Vault.
// On-chain identity resolves lazily via AgentIdentityResolver; ENS publishing
// is separate.
type BootstrapRunner struct {
	logger        *zap.Logger
	walletStorage identity_interfaces.WalletCredentialsStorageInterface
}

func NewBootstrapRunner(
	logger *zap.Logger,
	walletStorage identity_interfaces.WalletCredentialsStorageInterface,
) *BootstrapRunner {
	return &BootstrapRunner{
		logger:        logger,
		walletStorage: walletStorage,
	}
}

// Run loads (or generates) the wallet. Vault-backed generation is fast, so
// callers may run this synchronously at startup.
func (r *BootstrapRunner) Run(ctx context.Context) error {
	cred, err := r.walletStorage.GetOrGenerate(ctx)
	if err != nil {
		msg := fmt.Sprintf("wallet bootstrap failed: %v", err)
		r.logger.Error(msg)
		return err
	}

	r.logger.Info("wallet ready", zap.String("address", cred.Address))

	return nil
}
