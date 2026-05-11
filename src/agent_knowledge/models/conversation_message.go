package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationMessage persists one turn. Schema lives in runtime_db migrations.
// Dedup by Hash (SHA-512 of role/author/content/platform/conversationId);
// caller must demojize Content before write so reasoning + knowledge hash agree.
type ConversationMessage struct {
	Id             uuid.UUID `gorm:"primaryKey;size:36"`
	ConversationId string    `gorm:"not null;size:128;index:idx_conversation_messages_conversation_id"`
	Author         string    `gorm:"not null;size:128"`
	Role           string    `gorm:"not null;size:32"`
	Platform       string    `gorm:"size:64"`
	Content        string    `gorm:"not null;type:text"`
	Hash           string    `gorm:"not null;size:128;uniqueIndex:uq_conversation_messages_hash"`
	Timestamp      time.Time `gorm:"not null"`
	CreatedAt      time.Time `gorm:"not null"`
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
	hash, err := utils.ComputeConversationMessageHash(m.Role, m.Author, m.Platform, m.Content, m.ConversationId)
	if err != nil {
		return err
	}
	m.Hash = hash
	return nil
}
