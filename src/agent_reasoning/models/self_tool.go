package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SelfTool is the per-agent enabled flag for one built-in tool; rows are
// reconciled against the registry at runtime startup. Schema owned by
// runtime_db/migrations/.
type SelfTool struct {
	Id        uuid.UUID `gorm:"primaryKey;size:36"`
	AgentId   uuid.UUID `gorm:"not null;size:36;uniqueIndex:uq_self_tools_agent_name"`
	Name      string    `gorm:"not null;size:128;uniqueIndex:uq_self_tools_agent_name"`
	Enabled   bool      `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (SelfTool) TableName() string {
	return "self_tools"
}

func (t *SelfTool) BeforeCreate(_ *gorm.DB) error {
	if t.Id == uuid.Nil {
		t.Id = uuid.New()
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	return nil
}

func (t *SelfTool) BeforeSave(_ *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}
