package models

import (
	"errors"
	"math/big"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentSpendRule bounds a spend in atomic units; "0" denies, and there is no unlimited sentinel.
type PaymentSpendRule struct {
	Id       uuid.UUID           `gorm:"primaryKey;size:36"`
	PolicyId uuid.UUID           `gorm:"not null;size:36;index"`
	Policy   PaymentSpendPolicy  `gorm:"foreignKey:PolicyId"`
	Position int                 `gorm:"not null"`
	Name     string              `gorm:"not null;size:64"`
	Method   enums.PaymentMethod `gorm:"not null;size:32"`

	MaxAmountPerPayment string `gorm:"not null;size:78"`
	// 0 seconds leaves the rule bounded per payment only.
	WindowSeconds      int    `gorm:"not null;default:0"`
	MaxAmountPerWindow string `gorm:"not null;size:78;default:0"`
	// Runtime state, never operator input; the period rolls over on first use after it expires.
	WindowSpent     string     `gorm:"not null;size:78;default:0"`
	WindowStartedAt *time.Time `gorm:""`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// SpentInWindow is what the next payment is measured against; an elapsed period reads as nothing spent.
func (m *PaymentSpendRule) SpentInWindow(now time.Time) string {
	if m.WindowSeconds <= 0 || m.windowExpiredAt(now) {
		return "0"
	}
	return m.WindowSpent
}

// WindowResetsAt is nil while no period is running; an elapsed one restarts on first use.
func (m *PaymentSpendRule) WindowResetsAt(now time.Time) *time.Time {
	if m.WindowSeconds <= 0 || m.WindowStartedAt == nil || m.windowExpiredAt(now) {
		return nil
	}
	resetsAt := m.WindowStartedAt.Add(time.Duration(m.WindowSeconds) * time.Second)
	return &resetsAt
}

// RecordWindowSpend rolls an elapsed period and adds the amount, reporting false when it cannot absorb it.
func (m *PaymentSpendRule) RecordWindowSpend(amount *big.Int, now time.Time) (bool, error) {
	if m.WindowSeconds <= 0 {
		return true, nil
	}

	budget, ok := new(big.Int).SetString(m.MaxAmountPerWindow, 10)
	if !ok || budget.Sign() < 0 {
		return false, errors.New("window budget is not a usable amount")
	}

	spent := new(big.Int)
	if !m.windowExpiredAt(now) {
		if spent, ok = new(big.Int).SetString(m.WindowSpent, 10); !ok || spent.Sign() < 0 {
			return false, errors.New("window spend is not a usable amount")
		}
	} else {
		started := now
		m.WindowStartedAt = &started
	}

	after := new(big.Int).Add(spent, amount)
	if after.Cmp(budget) > 0 {
		return false, nil
	}
	m.WindowSpent = after.String()
	return true, nil
}

// CoversPayment reports whether the rule's per-payment cap allows the amount.
func (m *PaymentSpendRule) CoversPayment(amount *big.Int) (bool, error) {
	cap, ok := new(big.Int).SetString(m.MaxAmountPerPayment, 10)
	if !ok || cap.Sign() < 0 {
		return false, errors.New("cap is not a usable amount")
	}
	return amount.Cmp(cap) <= 0, nil
}

func (m *PaymentSpendRule) windowExpiredAt(now time.Time) bool {
	if m.WindowStartedAt == nil {
		return true
	}
	return !now.Before(m.WindowStartedAt.Add(time.Duration(m.WindowSeconds) * time.Second))
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
