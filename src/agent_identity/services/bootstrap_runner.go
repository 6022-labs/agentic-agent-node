package services

import (
	"context"
	"errors"
	"fmt"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"go.uber.org/zap"
)

// BootstrapRunner loads or generates the agent's wallet in the local Vault, then
// best-effort refreshes the agent's ENS runtime records. On-chain identity
// resolves lazily via AgentIdentityResolver.
type BootstrapRunner struct {
	logger                   *zap.Logger
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface
	ensPublisher             identity_interfaces.EnsPublisherInterface
}

func NewBootstrapRunner(
	logger *zap.Logger,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
	ensPublisher identity_interfaces.EnsPublisherInterface,
) *BootstrapRunner {
	return &BootstrapRunner{
		logger:                   logger,
		walletCredentialsStorage: walletCredentialsStorage,
		ensPublisher:             ensPublisher,
	}
}

// Run loads (or generates) the wallet synchronously (Vault-backed, fast), then
// kicks off the ENS runtime-records refresh in the background.
func (r *BootstrapRunner) Run(ctx context.Context) error {
	cred, err := r.walletCredentialsStorage.GetOrGenerate(ctx)
	if err != nil {
		msg := fmt.Sprintf("wallet bootstrap failed: %v", err)
		r.logger.Error(msg)
		return err
	}

	r.logger.Info("wallet ready", zap.String("address", cred.Address))

	// Best-effort, non-blocking: refresh ENS records if minted (e.g. URL changed).
	go r.refreshEnsRuntimeRecords(ctx)

	return nil
}

func (r *BootstrapRunner) refreshEnsRuntimeRecords(ctx context.Context) {
	err := r.ensPublisher.PublishRuntimeRecords(ctx)
	if err == nil {
		return
	}
	if errors.Is(err, common_interfaces.ErrIdentityNotFound) {
		r.logger.Info("skipping ENS runtime records refresh: agent not minted yet")
		return
	}
	r.logger.Warn("ENS runtime records refresh failed; will retry next boot", zap.Error(err))
}
