package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MCPConnectionRepository struct {
	client *runtime_db.DatabaseClient
}

func NewMCPConnectionRepository(client *runtime_db.DatabaseClient) *MCPConnectionRepository {
	return &MCPConnectionRepository{client: client}
}

func (r *MCPConnectionRepository) List(ctx context.Context) ([]models.MCPConnection, error) {
	var entities []models.MCPConnection
	if err := r.client.GetDB().WithContext(ctx).Order("created_at ASC").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *MCPConnectionRepository) GetById(ctx context.Context, id uuid.UUID) (*models.MCPConnection, error) {
	var entity models.MCPConnection
	err := r.client.GetDB().WithContext(ctx).First(&entity, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *MCPConnectionRepository) Create(ctx context.Context, connection *models.MCPConnection) (*models.MCPConnection, error) {
	if err := r.client.GetDB().WithContext(ctx).Create(connection).Error; err != nil {
		return nil, err
	}
	return connection, nil
}

func (r *MCPConnectionRepository) Update(ctx context.Context, connection *models.MCPConnection) (*models.MCPConnection, error) {
	if err := r.client.GetDB().WithContext(ctx).Save(connection).Error; err != nil {
		return nil, err
	}
	return connection, nil
}

func (r *MCPConnectionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.GetDB().WithContext(ctx).Delete(&models.MCPConnection{}, "id = ?", id).Error
}
