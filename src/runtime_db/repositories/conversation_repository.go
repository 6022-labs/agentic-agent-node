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

func (r *ConversationRepository) Create(
	ctx context.Context,
	conversation *models.Conversation,
) (*models.Conversation, error) {
	if err := r.client.GetDB().WithContext(ctx).Create(conversation).Error; err != nil {
		return nil, err
	}
	return conversation, nil
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

func (r *ConversationRepository) GetByIdAndAgentIdAndCallerId(
	ctx context.Context,
	id string,
	agentId uuid.UUID,
	callerId *string,
) (*models.Conversation, error) {
	var entity models.Conversation
	query := r.client.GetDB().WithContext(ctx).Where("id = ? AND agent_id = ?", id, agentId)
	err := r.scopeToCallerId(query, callerId).First(&entity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *ConversationRepository) ExistsByAgentIdAndCallerIdAndName(
	ctx context.Context,
	agentId uuid.UUID,
	callerId *string,
	name string,
	excludedId *string,
) (bool, error) {
	query := r.client.GetDB().
		WithContext(ctx).
		Model(&models.Conversation{}).
		Where("agent_id = ? AND name = ?", agentId, name)

	if callerId != nil {
		query = query.Where("caller_id = ?", *callerId)
	} else {
		query = query.Where("caller_id IS NULL")
	}

	if excludedId != nil {
		query = query.Where("id <> ?", *excludedId)
	}

	var total int64
	if err := query.Limit(1).Count(&total).Error; err != nil {
		return false, err
	}
	return total > 0, nil
}

func (r *ConversationRepository) Update(
	ctx context.Context,
	conversation *models.Conversation,
) (*models.Conversation, error) {
	if err := r.client.GetDB().WithContext(ctx).Save(conversation).Error; err != nil {
		return nil, err
	}
	return conversation, nil
}

func (r *ConversationRepository) ListByAgentId(
	ctx context.Context,
	agentId uuid.UUID,
	callerId *string,
	offset *int,
	limit *int,
) ([]models.Conversation, error) {
	query := r.scopeToCallerId(r.client.GetDB().WithContext(ctx).Where("agent_id = ?", agentId), callerId).
		Order("created_at DESC")

	if limit != nil {
		query = query.Limit(*limit)
	}
	if offset != nil {
		query = query.Offset(*offset)
	}

	var entities []models.Conversation
	if err := query.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *ConversationRepository) CountByAgentId(
	ctx context.Context,
	agentId uuid.UUID,
	callerId *string,
) (uint, error) {
	var total int64
	query := r.client.GetDB().WithContext(ctx).Model(&models.Conversation{}).Where("agent_id = ?", agentId)
	if err := r.scopeToCallerId(query, callerId).Count(&total).Error; err != nil {
		return 0, err
	}
	return uint(total), nil
}

// scopeToCallerId: nil leaves the query unscoped.
func (r *ConversationRepository) scopeToCallerId(query *gorm.DB, callerId *string) *gorm.DB {
	if callerId == nil {
		return query
	}
	return query.Where("caller_id = ?", *callerId)
}
