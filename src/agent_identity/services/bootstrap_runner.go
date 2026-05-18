package services

import (
	"context"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"go.uber.org/zap"
)

// BootstrapRunner ensures the agent's wallet is loaded from (or generated
// into) the local Vault, then publishes the agent's runtime ENS records.
// Wallet readiness is required for the gateway to accept traffic; on-chain
// identity is resolved lazily by AgentIdentityResolver as routes need it.
// ENS publish runs after the readiness signal — failures are operational
// (RPC down, NFT not yet minted, chain congested) and must not crash the
// node or block readiness.
type BootstrapRunner struct {
	logger        *zap.Logger
	walletStorage identity_interfaces.WalletCredentialsStorageInterface
	stateHolder   identity_interfaces.BootstrapStateHolderInterface
	ensPublisher  identity_interfaces.EnsRuntimeRecordsPublisherInterface
}

func NewBootstrapRunner(
	logger *zap.Logger,
	walletStorage identity_interfaces.WalletCredentialsStorageInterface,
	stateHolder identity_interfaces.BootstrapStateHolderInterface,
	ensPublisher identity_interfaces.EnsRuntimeRecordsPublisherInterface,
) *BootstrapRunner {
	return &BootstrapRunner{
		logger:        logger,
		walletStorage: walletStorage,
		stateHolder:   stateHolder,
		ensPublisher:  ensPublisher,
	}
}

// Run loads (or generates) the wallet, moves the holder to ready, then
// publishes the agent's runtime ENS records. ENS failure is non-fatal.
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

	if err := r.ensPublisher.PublishRuntimeRecords(ctx); err != nil {
		r.logger.Warn("ENS runtime records publish failed", zap.Error(err))
	} else {
		r.logger.Info("ENS runtime records published")
	}

	return nil
}
