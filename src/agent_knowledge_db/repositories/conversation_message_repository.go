package repositories

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db"
	"gorm.io/gorm/clause"
)

type ConversationMessageRepository struct {
	client *agent_knowledge_db.DatabaseClient
}

func NewConversationMessageRepository(client *agent_knowledge_db.DatabaseClient) *ConversationMessageRepository {
	return &ConversationMessageRepository{client: client}
}

// AppendMessages computes a hash for any input that lacks one, drops rows
// whose hash already exists, and inserts the rest. The unique index on `hash`
// is the source of truth — `ON CONFLICT (hash) DO NOTHING` covers concurrent
// writers.
func (r *ConversationMessageRepository) AppendMessages(messages []models.ConversationMessage) ([]models.ConversationMessage, error) {
	if len(messages) == 0 {
		return nil, nil
	}

	hashes := make([]string, 0, len(messages))
	for i := range messages {
		if messages[i].Hash == "" {
			if err := messages[i].ComputeHash(); err != nil {
				return nil, err
			}
		}
		hashes = append(hashes, messages[i].Hash)
	}

	var existing []string
	if err := r.client.GetDB().
		Model(&models.ConversationMessage{}).
		Where("hash IN ?", hashes).
		Pluck("hash", &existing).Error; err != nil {
		return nil, err
	}
	existingSet := make(map[string]struct{}, len(existing))
	for _, h := range existing {
		existingSet[h] = struct{}{}
	}

	toInsert := make([]models.ConversationMessage, 0, len(messages))
	for _, m := range messages {
		if _, hit := existingSet[m.Hash]; hit {
			continue
		}
		toInsert = append(toInsert, m)
	}
	if len(toInsert) == 0 {
		return nil, nil
	}

	if err := r.client.GetDB().
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "hash"}}, DoNothing: true}).
		Create(&toInsert).Error; err != nil {
		return nil, err
	}
	return toInsert, nil
}

func (r *ConversationMessageRepository) ListByConversation(conversationId string) ([]models.ConversationMessage, error) {
	var entities []models.ConversationMessage
	if err := r.client.GetDB().
		Where("conversation_id = ?", conversationId).
		Order("timestamp ASC").
		Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}
