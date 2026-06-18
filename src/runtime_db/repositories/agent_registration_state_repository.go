package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"gorm.io/gorm"
)

type AgentRegistrationStateRepository struct {
	client *runtime_db.DatabaseClient
}

func NewAgentRegistrationStateRepository(client *runtime_db.DatabaseClient) *AgentRegistrationStateRepository {
	return &AgentRegistrationStateRepository{client: client}
}

func (r *AgentRegistrationStateRepository) Get(ctx context.Context) (*models.AgentRegistrationState, error) {
	var entity models.AgentRegistrationState
	err := r.client.GetDB().WithContext(ctx).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *AgentRegistrationStateRepository) Upsert(ctx context.Context, state *models.AgentRegistrationState) (*models.AgentRegistrationState, error) {
	existing, err := r.Get(ctx)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if err := r.client.GetDB().WithContext(ctx).Create(state).Error; err != nil {
			return nil, err
		}
		return state, nil
	}

	existing.MintStatus = state.MintStatus
	existing.MintChainId = state.MintChainId
	existing.MintTxHash = state.MintTxHash
	existing.MintedAt = state.MintedAt
	existing.EnsRegistryStatus = state.EnsRegistryStatus
	existing.EnsChainId = state.EnsChainId
	existing.EnsTxHash = state.EnsTxHash
	existing.EnsPublishedAt = state.EnsPublishedAt
	if err := r.client.GetDB().WithContext(ctx).Save(existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}
