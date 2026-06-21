package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_common/utils"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

// ConversationMessage persists one discussion turn — text, an assistant tool-call turn (ToolCalls), or a tool result (ToolCallId); dedup by Hash, Content demojized before write so reasoning + knowledge agree.
type ConversationMessage struct {
	Id             uuid.UUID                     `gorm:"primaryKey;size:36"`
	ConversationId string                        `gorm:"not null;size:128;index:idx_conversation_messages_conversation_id"`
	Conversation   Conversation                  `gorm:"foreignKey:ConversationId;references:Id;constraint:OnDelete:CASCADE"`
	Role           enums.ConversationMessageRole `gorm:"not null;size:32"`
	Content        *string                       `gorm:"type:text"`
	ToolCalls      *string                       `gorm:"type:text"`
	ToolCallId     *string                       `gorm:"size:128"`
	Hash           string                        `gorm:"not null;size:128;uniqueIndex:uq_conversation_messages_hash"`
	Timestamp      time.Time                     `gorm:"not null"`
	CreatedAt      time.Time                     `gorm:"not null"`
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
	hash, err := utils.ComputeConversationMessageHash(string(m.Role), lo.FromPtr(m.Content), lo.FromPtr(m.ToolCalls), lo.FromPtr(m.ToolCallId), m.ConversationId)
	if err != nil {
		return err
	}
	m.Hash = hash
	return nil
}
