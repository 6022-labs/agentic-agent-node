package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationMessage persists one turn. Agent scope is transitive via
// Conversation (conversation_id FK), mirroring payment_rule -> payment_policy.
// Dedup by Hash (SHA-512 of role/content/conversationId); caller must
// demojize Content before write so reasoning + knowledge hash agree.
type ConversationMessage struct {
	Id             uuid.UUID    `gorm:"primaryKey;size:36"`
	ConversationId string       `gorm:"not null;size:128;index:idx_conversation_messages_conversation_id"`
	Conversation   Conversation `gorm:"foreignKey:ConversationId;references:Id;constraint:OnDelete:CASCADE"`
	Role           string       `gorm:"not null;size:32"`
	Content        string       `gorm:"not null;type:text"`
	Hash           string       `gorm:"not null;size:128;uniqueIndex:uq_conversation_messages_hash"`
	Timestamp      time.Time    `gorm:"not null"`
	CreatedAt      time.Time    `gorm:"not null"`
}

func (ConversationMessage) TableName() string {
	return "conversation_messages"
}

func (m *ConversationMessage) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.Hash == "" {
		return m.ComputeHash()
	}
	return nil
}

func (m *ConversationMessage) ComputeHash() error {
	hash, err := utils.ComputeConversationMessageHash(m.Role, m.Content, m.ConversationId)
	if err != nil {
		return err
	}
	m.Hash = hash
	return nil
}
