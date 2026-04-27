package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MCPConnection is one MCP server the agent can call into. Schema is owned by
// the goose migration in runtime_db/migrations/. An agent may have many.
//
// HeaderAuthKeys lists the header names the connection authenticates with;
// the values live in Vault and are loaded into HeaderAuth on read.
type MCPConnection struct {
	Id             uuid.UUID `gorm:"primaryKey"`
	Name           string
	Url            string
	HeaderAuthKeys MCPConnectionHeaderAuthKeys
	CreatedAt      time.Time
	UpdatedAt      time.Time

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
