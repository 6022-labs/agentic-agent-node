package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LLMConnectionRepository struct {
	client *runtime_db.DatabaseClient
}

func NewLLMConnectionRepository(client *runtime_db.DatabaseClient) *LLMConnectionRepository {
	return &LLMConnectionRepository{client: client}
}

func (r *LLMConnectionRepository) GetByAgentId(ctx context.Context, agentId uuid.UUID) (*models.LLMConnection, error) {
	var entity models.LLMConnection
	err := r.client.GetDB().WithContext(ctx).First(&entity, "agent_id = ?", agentId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// Upsert keeps a single row per agent: update in place if one exists, otherwise insert.
func (r *LLMConnectionRepository) Upsert(ctx context.Context, connection *models.LLMConnection) (*models.LLMConnection, error) {
	existing, err := r.GetByAgentId(ctx, connection.AgentId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		if err := r.client.GetDB().WithContext(ctx).Create(connection).Error; err != nil {
			return nil, err
		}
		return connection, nil
	}

	existing.Platform = connection.Platform
	existing.Model = connection.Model
	existing.BaseUrl = connection.BaseUrl
	existing.SetupStatus = connection.SetupStatus
	if err := r.client.GetDB().WithContext(ctx).Save(existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

func (r *LLMConnectionRepository) DeleteByAgentId(ctx context.Context, agentId uuid.UUID) error {
	return r.client.GetDB().
		WithContext(ctx).
		Where("agent_id = ?", agentId).
		Delete(&models.LLMConnection{}).Error
}
