package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Conversation is a transcript head; a nil CallerId marks a legacy head only the owner sees.
type Conversation struct {
	Id        string    `gorm:"primaryKey;size:128"`
	AgentId   uuid.UUID `gorm:"not null;size:36;index;uniqueIndex:uq_conversations_agent_caller_name,priority:1"`
	CallerId  *string   `gorm:"size:128;uniqueIndex:uq_conversations_agent_caller_name,priority:2"`
	Name      *string   `gorm:"size:128;uniqueIndex:uq_conversations_agent_caller_name,priority:3"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Conversation) TableName() string {
	return "conversations"
}

func (m *Conversation) BeforeCreate(_ *gorm.DB) error {
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *Conversation) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
