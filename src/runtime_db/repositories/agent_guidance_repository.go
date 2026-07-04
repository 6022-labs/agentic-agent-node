package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AgentGuidanceRepository struct {
	client *runtime_db.DatabaseClient
}

func NewAgentGuidanceRepository(client *runtime_db.DatabaseClient) *AgentGuidanceRepository {
	return &AgentGuidanceRepository{client: client}
}

func (r *AgentGuidanceRepository) GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.AgentGuidance, error) {
	var entity models.AgentGuidance
	err := r.client.GetDB().WithContext(ctx).First(&entity, "agent_id = ?", agentId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *AgentGuidanceRepository) Upsert(ctx context.Context, guidance *models.AgentGuidance) (*models.AgentGuidance, error) {
	existing, err := r.GetByAgentId(ctx, guidance.AgentId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if err := r.client.GetDB().WithContext(ctx).Create(guidance).Error; err != nil {
			return nil, err
		}
		return guidance, nil
	}

	existing.Content = guidance.Content
	if err := r.client.GetDB().WithContext(ctx).Save(existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}
