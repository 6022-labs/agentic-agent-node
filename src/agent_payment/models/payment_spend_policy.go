package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentSpendPolicy is the head of what the agent may pay out; one row per agent.
type PaymentSpendPolicy struct {
	Id        uuid.UUID `gorm:"primaryKey;size:36"`
	AgentId   uuid.UUID `gorm:"not null;size:36;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (PaymentSpendPolicy) TableName() string {
	return "payment_spend_policies"
}

func (m *PaymentSpendPolicy) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *PaymentSpendPolicy) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
