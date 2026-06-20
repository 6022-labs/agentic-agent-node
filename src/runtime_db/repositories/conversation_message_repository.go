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

func (r *ConversationMessageRepository) AppendMessages(ctx context.Context, messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
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

func (r *ConversationMessageRepository) ListByConversationId(ctx context.Context, conversationId string) ([]models.ConversationMessage, error) {
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
