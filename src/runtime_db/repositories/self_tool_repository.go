package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SelfToolRepository struct {
	client *runtime_db.DatabaseClient
}

func NewSelfToolRepository(client *runtime_db.DatabaseClient) *SelfToolRepository {
	return &SelfToolRepository{client: client}
}

func (r *SelfToolRepository) ListByAgentId(ctx context.Context, agentId uuid.UUID) ([]models.SelfTool, error) {
	var entities []models.SelfTool
	if err := r.client.GetDB().WithContext(ctx).Where("agent_id = ?", agentId).Order("name ASC").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *SelfToolRepository) GetById(ctx context.Context, agentId uuid.UUID, id uuid.UUID) (*models.SelfTool, error) {
	var entity models.SelfTool
	err := r.client.GetDB().WithContext(ctx).First(&entity, "agent_id = ? AND id = ?", agentId, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *SelfToolRepository) Create(ctx context.Context, tool *models.SelfTool) (*models.SelfTool, error) {
	if err := r.client.GetDB().WithContext(ctx).Create(tool).Error; err != nil {
		return nil, err
	}
	return tool, nil
}

func (r *SelfToolRepository) Update(ctx context.Context, tool *models.SelfTool) (*models.SelfTool, error) {
	if err := r.client.GetDB().WithContext(ctx).Save(tool).Error; err != nil {
		return nil, err
	}
	return tool, nil
}

func (r *SelfToolRepository) ListExistingNames(ctx context.Context, agentId uuid.UUID, names []string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}
	var existing []string
	if err := r.client.GetDB().WithContext(ctx).Model(&models.SelfTool{}).
		Where("agent_id = ? AND name IN ?", agentId, names).
		Pluck("name", &existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

func (r *SelfToolRepository) DeleteAllExceptNames(ctx context.Context, agentId uuid.UUID, names []string) (int64, error) {
	db := r.client.GetDB().WithContext(ctx)
	// An empty NOT IN would match nothing in SQL; an empty registry means every row is stale.
	if len(names) == 0 {
		result := db.Delete(&models.SelfTool{}, "agent_id = ?", agentId)
		return result.RowsAffected, result.Error
	}
	result := db.Delete(&models.SelfTool{}, "agent_id = ? AND name NOT IN ?", agentId, names)
	return result.RowsAffected, result.Error
}
