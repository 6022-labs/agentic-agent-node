package services

import (
	"context"
	"fmt"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	identity_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_identity/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// SelfToolReconciler syncs self_tools rows with the registry once at runtime startup.
type SelfToolReconciler struct {
	logger             *zap.Logger
	selfToolRegistry   interfaces.SelfToolRegistryInterface
	selfToolRepository repositories.SelfToolRepositoryInterface
	agentResolver      identity_interfaces.AgentResolverInterface
}

func NewSelfToolReconciler(
	logger *zap.Logger,
	selfToolRegistry interfaces.SelfToolRegistryInterface,
	selfToolRepository repositories.SelfToolRepositoryInterface,
	agentResolver identity_interfaces.AgentResolverInterface,
) *SelfToolReconciler {
	return &SelfToolReconciler{
		logger:             logger,
		selfToolRegistry:   selfToolRegistry,
		selfToolRepository: selfToolRepository,
		agentResolver:      agentResolver,
	}
}

func (r *SelfToolReconciler) Reconcile(ctx context.Context) error {
	agent, err := r.agentResolver.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to load agent: %w", err)
	}
	if agent == nil {
		return apperrors.ErrAgentNotInitialized
	}
	agentId := agent.Id

	defaultEnabled := r.selfToolRegistry.ListDefaultEnabled()
	names := lo.Keys(defaultEnabled)

	existingNames, err := r.selfToolRepository.ListExistingNames(ctx, agentId, names)
	if err != nil {
		return fmt.Errorf("failed to list existing self tool names: %w", err)
	}

	for name, enabled := range defaultEnabled {
		if lo.Contains(existingNames, name) {
			continue
		}
		if _, err := r.selfToolRepository.Create(ctx, &models.SelfTool{
			AgentId: agentId,
			Name:    name,
			Enabled: enabled,
		}); err != nil {
			return fmt.Errorf("failed to create self tool row %q: %w", name, err)
		}
		r.logger.Info("self tool registered", zap.String("toolName", name), zap.Bool("enabled", enabled))
	}

	deleted, err := r.selfToolRepository.DeleteAllExceptNames(ctx, agentId, names)
	if err != nil {
		return fmt.Errorf("failed to delete stale self tool rows: %w", err)
	}
	if deleted > 0 {
		r.logger.Info("stale self tools removed", zap.Int64("count", deleted))
	}

	return nil
}
