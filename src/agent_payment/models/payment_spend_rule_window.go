package models

import (
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentSpendRuleWindow caps a rule's total spend over a recurring period; one per rule and period length.
type PaymentSpendRuleWindow struct {
	Id                 uuid.UUID        `gorm:"primaryKey;size:36"`
	PaymentSpendRuleId uuid.UUID        `gorm:"not null;size:36;uniqueIndex:uq_payment_spend_rule_windows_rule_seconds,priority:1"`
	PaymentSpendRule   PaymentSpendRule `gorm:"foreignKey:PaymentSpendRuleId;references:Id;constraint:OnDelete:CASCADE"`
	WindowSeconds      int              `gorm:"not null;uniqueIndex:uq_payment_spend_rule_windows_rule_seconds,priority:2"`
	MaxAmount          string           `gorm:"not null;size:78"`
	// Runtime state, never operator input; the period rolls over on first use after it expires.
	Spent     string     `gorm:"not null;size:78;default:0"`
	StartedAt *time.Time `gorm:""`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

// SpentAt is what the next payment is measured against; an elapsed period reads as nothing spent.
func (m *PaymentSpendRuleWindow) SpentAt(now time.Time) string {
	if m.expiredAt(now) {
		return "0"
	}
	return m.Spent
}

// ResetsAt is nil while no period is running; an elapsed one restarts on first use.
func (m *PaymentSpendRuleWindow) ResetsAt(now time.Time) *time.Time {
	if m.expiredAt(now) {
		return nil
	}
	resetsAt := m.StartedAt.Add(time.Duration(m.WindowSeconds) * time.Second)
	return &resetsAt
}

// RecordSpend rolls an elapsed period and adds the amount, reporting false when the budget cannot absorb it.
func (m *PaymentSpendRuleWindow) RecordSpend(amount *big.Int, now time.Time) (bool, error) {
	maxAmount, ok := new(big.Int).SetString(m.MaxAmount, 10)
	if !ok || maxAmount.Sign() < 0 {
		return false, errors.New("window budget is not a usable amount")
	}

	spent := new(big.Int)
	if !m.expiredAt(now) {
		if spent, ok = new(big.Int).SetString(m.Spent, 10); !ok || spent.Sign() < 0 {
			return false, errors.New("window spend is not a usable amount")
		}
	} else {
		// Whole seconds survive every driver's datetime precision, so Release can match the period exactly.
		started := now.Truncate(time.Second)
		m.StartedAt = &started
	}

	after := new(big.Int).Add(spent, amount)
	if after.Cmp(maxAmount) > 0 {
		return false, nil
	}
	m.Spent = after.String()
	return true, nil
}

// Release takes back an amount charged in the period started at periodStartedAt; a later period is left alone.
func (m *PaymentSpendRuleWindow) Release(amount *big.Int, periodStartedAt time.Time) error {
	if m.StartedAt == nil || !m.StartedAt.Equal(periodStartedAt) {
		return nil
	}
	spent, ok := new(big.Int).SetString(m.Spent, 10)
	if !ok || spent.Sign() < 0 {
		return errors.New("window spend is not a usable amount")
	}
	spent.Sub(spent, amount)
	if spent.Sign() < 0 {
		spent.SetInt64(0)
	}
	m.Spent = spent.String()
	return nil
}

func (m *PaymentSpendRuleWindow) Reset() {
	m.Spent = "0"
	m.StartedAt = nil
}

func (m *PaymentSpendRuleWindow) expiredAt(now time.Time) bool {
	if m.StartedAt == nil {
		return true
	}
	return !now.Before(m.StartedAt.Add(time.Duration(m.WindowSeconds) * time.Second))
}

func (PaymentSpendRuleWindow) TableName() string {
	return "payment_spend_rule_windows"
}

func (m *PaymentSpendRuleWindow) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *PaymentSpendRuleWindow) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
