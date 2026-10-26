package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentReceivePolicy is the head of the agent's pricing table; PaymentReceiveRule rows
// reference it. One row per agent; AgentId keeps the schema multi-agent
// ready. Schema owned by runtime_db/migrations/.
type PaymentReceivePolicy struct {
	Id        uuid.UUID `gorm:"primaryKey;size:36"`
	AgentId   uuid.UUID `gorm:"not null;size:36;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (PaymentReceivePolicy) TableName() string {
	return "payment_receive_policies"
}

func (m *PaymentReceivePolicy) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *PaymentReceivePolicy) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
