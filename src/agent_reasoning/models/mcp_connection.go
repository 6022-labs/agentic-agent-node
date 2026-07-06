package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MCPConnection is one MCP server the agent can call into (an agent may have
// many). HeaderAuthKeys names the auth headers; their values live in Vault and
// load into HeaderAuth on read. Schema owned by runtime_db/migrations/.
type MCPConnection struct {
	Id             uuid.UUID                   `gorm:"primaryKey;size:36"`
	AgentId        uuid.UUID                   `gorm:"not null;size:36;uniqueIndex:uq_mcp_connections_agent_name"`
	Name           string                      `gorm:"not null;size:128;uniqueIndex:uq_mcp_connections_agent_name"`
	Url            string                      `gorm:"not null;type:text"`
	HeaderAuthKeys MCPConnectionHeaderAuthKeys `gorm:"not null"`
	CreatedAt      time.Time                   `gorm:"not null"`
	UpdatedAt      time.Time                   `gorm:"not null"`

	HeaderAuth map[string]string `gorm:"-"`
}

func (MCPConnection) TableName() string {
	return "mcp_connections"
}

func (m *MCPConnection) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *MCPConnection) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
