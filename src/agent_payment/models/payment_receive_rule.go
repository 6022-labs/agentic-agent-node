package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentReceiveRule prices one token, named by its method's config row.
type PaymentReceiveRule struct {
	Id        uuid.UUID            `gorm:"primaryKey;size:36"`
	PolicyId  uuid.UUID            `gorm:"not null;size:36;index"`
	Policy    PaymentReceivePolicy `gorm:"foreignKey:PolicyId"`
	Name      string               `gorm:"not null;size:64"`
	Method    enums.PaymentMethod  `gorm:"not null;size:32"`
	CreatedAt time.Time            `gorm:"not null"`
	UpdatedAt time.Time            `gorm:"not null"`
}

func (PaymentReceiveRule) TableName() string {
	return "payment_receive_rules"
}

func (m *PaymentReceiveRule) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *PaymentReceiveRule) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
