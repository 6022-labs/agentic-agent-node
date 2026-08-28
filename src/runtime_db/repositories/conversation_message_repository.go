package repositories

import (
	"context"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"gorm.io/gorm/clause"
)

type ConversationMessageRepository struct {
	client *runtime_db.DatabaseClient
}

func NewConversationMessageRepository(client *runtime_db.DatabaseClient) *ConversationMessageRepository {
	return &ConversationMessageRepository{client: client}
}

func (r *ConversationMessageRepository) AppendMessages(
	ctx context.Context,
	messages []models.ConversationMessage,
) ([]models.ConversationMessage, error) {
	if len(messages) == 0 {
		return nil, nil
	}
	if err := r.client.GetDB().
		WithContext(ctx).
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "hash"}}, DoNothing: true}).
		Create(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *ConversationMessageRepository) ListByConversationId(
	ctx context.Context,
	conversationId string,
) ([]models.ConversationMessage, error) {
	var entities []models.ConversationMessage
	if err := r.client.GetDB().
		WithContext(ctx).
		Where("conversation_id = ?", conversationId).
		Order("timestamp ASC").
		Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *ConversationMessageRepository) ListExistingHashes(ctx context.Context, hashes []string) ([]string, error) {
	if len(hashes) == 0 {
		return nil, nil
	}
	var existing []string
	if err := r.client.GetDB().
		WithContext(ctx).
		Model(&models.ConversationMessage{}).
		Where("hash IN ?", hashes).
		Pluck("hash", &existing).Error; err != nil {
		return nil, err
	}
	return existing, nil
}

// CountByConversationIds groups on its own table; ids without a message are simply absent from the map.
func (r *ConversationMessageRepository) CountByConversationIds(
	ctx context.Context,
	conversationIds []string,
) (map[string]int64, error) {
	counts := make(map[string]int64, len(conversationIds))
	if len(conversationIds) == 0 {
		return counts, nil
	}

	var rows []struct {
		ConversationId string
		Total          int64
	}
	if err := r.client.GetDB().
		WithContext(ctx).
		Model(&models.ConversationMessage{}).
		Select("conversation_id, COUNT(id) AS total").
		Where("conversation_id IN ?", conversationIds).
		Group("conversation_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.ConversationId] = row.Total
	}
	return counts, nil
}
