package models

import (
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentReceiveRulePrice is what payers matching AddressPatterns pay for the rule's token.
type PaymentReceiveRulePrice struct {
	Id                   uuid.UUID          `gorm:"primaryKey;size:36"`
	PaymentReceiveRuleId uuid.UUID          `gorm:"not null;size:36;uniqueIndex:uq_payment_receive_rule_prices_rule_price,priority:1"`
	PaymentReceiveRule   PaymentReceiveRule `gorm:"foreignKey:PaymentReceiveRuleId;references:Id;constraint:OnDelete:CASCADE"`
	// Atomic units; "0" is free and never settles.
	Price           string   `gorm:"not null;size:78;uniqueIndex:uq_payment_receive_rule_prices_rule_price,priority:2"`
	AddressPatterns []string `gorm:"serializer:json;not null"`

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (PaymentReceiveRulePrice) TableName() string {
	return "payment_receive_rule_prices"
}

// Matches reports whether this price is open to the address.
func (m *PaymentReceiveRulePrice) Matches(address string) bool {
	// The address is lowercased; patterns are matched as written, so uppercase-hex patterns never match.
	lowered := strings.ToLower(strings.TrimSpace(address))
	for _, pattern := range m.AddressPatterns {
		compiled, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			continue
		}
		if compiled.MatchString(lowered) {
			return true
		}
	}
	return false
}

func (m *PaymentReceiveRulePrice) IsFree() bool {
	value, ok := new(big.Int).SetString(m.Price, 10)
	return ok && value.Sign() == 0
}

func (m *PaymentReceiveRulePrice) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.AddressPatterns == nil {
		m.AddressPatterns = []string{}
	}
	return nil
}

func (m *PaymentReceiveRulePrice) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
