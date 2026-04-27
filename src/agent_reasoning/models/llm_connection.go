package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LLMConnection is the agent's configured LLM provider. Schema is owned by
// the goose migration in runtime_db/migrations/. One agent per container, so
// the table is expected to hold zero or one row at any time.
//
// The api key lives in Vault, not on this row.
type LLMConnection struct {
	Id          uuid.UUID                      `gorm:"primaryKey;size:36"`
	Platform    enums.LLMConnectionPlatform    `gorm:"not null;size:64"`
	Model       string                         `gorm:"not null;size:255"`
	BaseUrl     *string                        `gorm:"type:text"`
	SetupStatus enums.LLMConnectionSetupStatus `gorm:"not null;size:32"`
	CreatedAt   time.Time                      `gorm:"not null"`
	UpdatedAt   time.Time                      `gorm:"not null"`
}

func (LLMConnection) TableName() string {
	return "llm_connections"
}

func (m *LLMConnection) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *LLMConnection) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
