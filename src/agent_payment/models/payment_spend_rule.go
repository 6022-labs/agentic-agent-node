package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentSpendRule caps one payment in atomic units; "0" denies, and there is no unlimited sentinel.
type PaymentSpendRule struct {
	Id        uuid.UUID           `gorm:"primaryKey;size:36"`
	PolicyId  uuid.UUID           `gorm:"not null;size:36;index"`
	Policy    PaymentSpendPolicy  `gorm:"foreignKey:PolicyId"`
	Position  int                 `gorm:"not null"`
	Name      string              `gorm:"not null;size:64"`
	MaxAmount string              `gorm:"not null;size:78"`
	Method    enums.PaymentMethod `gorm:"not null;size:32"`
	CreatedAt time.Time           `gorm:"not null"`
	UpdatedAt time.Time           `gorm:"not null"`
}

func (PaymentSpendRule) TableName() string {
	return "payment_spend_rules"
}

func (m *PaymentSpendRule) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *PaymentSpendRule) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
