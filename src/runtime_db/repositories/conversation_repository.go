package repositories

import (
	"context"
	"errors"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConversationRepository struct {
	client *runtime_db.DatabaseClient
}

func NewConversationRepository(client *runtime_db.DatabaseClient) *ConversationRepository {
	return &ConversationRepository{client: client}
}

func (r *ConversationRepository) Ensure(ctx context.Context, conversation *models.Conversation) (bool, error) {
	result := r.client.GetDB().
		WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoNothing: true}).
		Create(conversation)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *ConversationRepository) GetById(ctx context.Context, id string) (*models.Conversation, error) {
	var entity models.Conversation
	err := r.client.GetDB().WithContext(ctx).First(&entity, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *ConversationRepository) ListByAgentId(ctx context.Context, agentId uuid.UUID) ([]models.Conversation, error) {
	var entities []models.Conversation
	if err := r.client.GetDB().
		WithContext(ctx).
		Where("agent_id = ?", agentId).
		Order("created_at DESC").
		Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}
