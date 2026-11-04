package models

import (
	"errors"
	"math/big"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentSpendRule bounds what the agent pays in one token; its method config names the token.
type PaymentSpendRule struct {
	Id       uuid.UUID           `gorm:"primaryKey;size:36"`
	PolicyId uuid.UUID           `gorm:"not null;size:36;index"`
	Policy   PaymentSpendPolicy  `gorm:"foreignKey:PolicyId"`
	Name     string              `gorm:"not null;size:64"`
	Method   enums.PaymentMethod `gorm:"not null;size:32"`
	// nil leaves payments bounded by the windows only; "0" denies anything priced.
	MaxAmountPerPayment *string `gorm:"size:78"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (m *PaymentSpendRule) CoversPayment(amount *big.Int) (bool, error) {
	if m.MaxAmountPerPayment == nil {
		return true, nil
	}
	maxAmount, ok := new(big.Int).SetString(*m.MaxAmountPerPayment, 10)
	if !ok || maxAmount.Sign() < 0 {
		return false, errors.New("max amount per payment is not a usable amount")
	}
	return amount.Cmp(maxAmount) <= 0, nil
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
