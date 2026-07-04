package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentGuidance is the owner-editable prompt layer; one agent per container, so zero or one row.
type AgentGuidance struct {
	Id        uuid.UUID `gorm:"primaryKey;size:36"`
	AgentId   uuid.UUID `gorm:"not null;size:36;uniqueIndex"`
	Content   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (AgentGuidance) TableName() string {
	return "agent_guidances"
}

func (m *AgentGuidance) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *AgentGuidance) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
