package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	common_interfaces "github.com/6022-labs/agentic-agent-node/src/common/interfaces"
	"go.uber.org/zap"
)

// BootstrapRunner ensures the agent row exists (so secret paths can be scoped by
// its id), loads or generates the wallet, then best-effort refreshes ENS records.
type BootstrapRunner struct {
	logger                   *zap.Logger
	agentRepository          repositories.AgentRepositoryInterface
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface
	ensPublisher             identity_interfaces.EnsPublisherInterface
}

func NewBootstrapRunner(
	logger *zap.Logger,
	agentRepository repositories.AgentRepositoryInterface,
	walletCredentialsStorage identity_interfaces.WalletCredentialsStorageInterface,
	ensPublisher identity_interfaces.EnsPublisherInterface,
) *BootstrapRunner {
	return &BootstrapRunner{
		logger:                   logger,
		agentRepository:          agentRepository,
		walletCredentialsStorage: walletCredentialsStorage,
		ensPublisher:             ensPublisher,
	}
}

func (r *BootstrapRunner) Run(ctx context.Context) error {
	// The agent row must exist before any secret op: its id scopes every secret path.
	if err := r.ensureAgent(ctx); err != nil {
		r.logger.Error(fmt.Sprintf("agent bootstrap failed: %v", err))
		return err
	}

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

// ensureAgent creates an empty agent draft if none exists; registration later fills the same row.
func (r *BootstrapRunner) ensureAgent(ctx context.Context) error {
	agent, err := r.agentRepository.Get(ctx)
	if err != nil {
		return err
	}
	if agent != nil {
		return nil
	}

	created, err := r.agentRepository.Upsert(ctx, &models.Agent{})
	if err != nil {
		return err
	}
	r.logger.Info("agent draft created", zap.String("id", created.Id.String()))
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
