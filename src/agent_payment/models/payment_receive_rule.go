package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentReceiveRule is method-neutral: atomic-unit Price ("0" = free) and Method
// names the payment method. The method's typed settings live in its own
// per-method config table (keyed by this rule's Id), owned by the method's
// strategy — the Method column is the discriminator. Position keeps table
// order. Schema owned by runtime_db/migrations/.
type PaymentReceiveRule struct {
	Id        uuid.UUID            `gorm:"primaryKey;size:36"`
	PolicyId  uuid.UUID            `gorm:"not null;size:36;index"`
	Policy    PaymentReceivePolicy `gorm:"foreignKey:PolicyId"`
	Position  int                  `gorm:"not null"`
	Name      string               `gorm:"not null;size:64"`
	Price     string               `gorm:"not null;size:78"`
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
