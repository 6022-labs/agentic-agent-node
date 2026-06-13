package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentRule prices one row of the policy: atomic-unit Price ("0" = free),
// AddressPatterns regexes full-matched on the lowercase caller address
// (".*" = everyone, the only rules advertised in 402s), Position keeps table
// order. Schema owned by runtime_db/migrations/.
type PaymentRule struct {
	Id                uuid.UUID     `gorm:"primaryKey;size:36"`
	PolicyId          uuid.UUID     `gorm:"not null;size:36;index"`
	Policy            PaymentPolicy `gorm:"foreignKey:PolicyId"`
	Position          int           `gorm:"not null"`
	Name              string        `gorm:"not null;size:64"`
	Network           string        `gorm:"not null;size:64"`
	Asset             string        `gorm:"not null;size:42"`
	PayTo             string        `gorm:"not null;size:42"`
	Price             string        `gorm:"not null;size:78"`
	MaxTimeoutSeconds int           `gorm:"not null"`
	Eip712Name        string        `gorm:"not null;size:64"`
	Eip712Version     string        `gorm:"not null;size:16"`
	AddressPatterns   []string      `gorm:"not null;serializer:json"`
	CreatedAt         time.Time     `gorm:"not null"`
	UpdatedAt         time.Time     `gorm:"not null"`
}

func (PaymentRule) TableName() string {
	return "payment_rules"
}

func (m *PaymentRule) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *PaymentRule) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
