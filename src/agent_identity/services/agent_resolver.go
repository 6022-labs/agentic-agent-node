package services

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"go.uber.org/zap"
)

const identitySyncTimeout = 30 * time.Second

// AgentResolver serves the agent row and refreshes its on-chain identity at most once per sync interval.
type AgentResolver struct {
	logger                 *zap.Logger
	agentRepository        repositories.AgentRepositoryInterface
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface
	agentIdentitySettings  *settings.AgentIdentitySettings

	inFlight   atomic.Bool
	lastSyncAt atomic.Int64
}

func NewAgentResolver(
	logger *zap.Logger,
	agentRepository repositories.AgentRepositoryInterface,
	registrationReconciler identity_interfaces.RegistrationReconcilerInterface,
	agentIdentitySettings *settings.AgentIdentitySettings,
) *AgentResolver {
	return &AgentResolver{
		logger:                 logger,
		agentRepository:        agentRepository,
		registrationReconciler: registrationReconciler,
		agentIdentitySettings:  agentIdentitySettings,
	}
}

func (r *AgentResolver) Get(ctx context.Context) (*models.Agent, error) {
	agent, err := r.agentRepository.Get(ctx)
	if err != nil || agent == nil {
		return agent, err
	}
	if agent.CollectionTokenId != nil {
		r.syncIfStale()
	}
	return agent, nil
}

// syncIfStale starts at most one background sync per interval; a failed attempt also waits for the next one.
func (r *AgentResolver) syncIfStale() {
	if !r.isStale() || !r.inFlight.CompareAndSwap(false, true) {
		return
	}
	// Re-checked under the flag: a sync may have finished between the first check and the swap.
	if !r.isStale() {
		r.inFlight.Store(false)
		return
	}
	r.lastSyncAt.Store(time.Now().UnixNano())

	// Background, not the request ctx: its values may carry an open unit-of-work transaction.
	syncCtx, cancel := context.WithTimeout(context.Background(), identitySyncTimeout)
	go func() {
		defer cancel()
		defer r.inFlight.Store(false)
		if err := r.registrationReconciler.ReconcileIdentity(syncCtx); err != nil {
			r.logger.Warn("agent identity sync failed", zap.Error(err))
		}
	}()
}

func (r *AgentResolver) isStale() bool {
	return time.Since(time.Unix(0, r.lastSyncAt.Load())) >= r.agentIdentitySettings.IdentitySyncInterval
}
