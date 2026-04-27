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
	Id             uuid.UUID                   `gorm:"primaryKey;size:36"`
	Name           string                      `gorm:"not null;size:128"`
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
