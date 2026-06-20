package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LLMConnection is the agent's configured LLM provider. One agent per
// container, so the table holds zero or one row. The api key lives in Vault,
// not here. Schema owned by runtime_db/migrations/.
type LLMConnection struct {
	Id          uuid.UUID                      `gorm:"primaryKey;size:36"`
	AgentId     uuid.UUID                      `gorm:"not null;size:36;uniqueIndex"`
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
