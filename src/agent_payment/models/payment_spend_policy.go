package models

import (
	"time"

	payment_helpers "github.com/6022-labs/agentic-agent-node/src/agent_payment/helpers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PaymentSpendPolicy is the head of what the agent may pay out; one row per agent.
type PaymentSpendPolicy struct {
	Id      uuid.UUID `gorm:"primaryKey;size:36"`
	AgentId uuid.UUID `gorm:"not null;size:36;uniqueIndex"`
	// Who the agent may pay, whatever the token; empty pays nobody.
	PayToPatterns []string  `gorm:"serializer:json;not null"`
	CreatedAt     time.Time `gorm:"not null"`
	UpdatedAt     time.Time `gorm:"not null"`
}

func (PaymentSpendPolicy) TableName() string {
	return "payment_spend_policies"
}

func (m *PaymentSpendPolicy) AllowsPayee(payTo string) bool {
	return payment_helpers.MatchesAddressPatterns(m.PayToPatterns, payTo)
}

func (m *PaymentSpendPolicy) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	// A nil slice serializes as null, which the not-null column refuses.
	if m.PayToPatterns == nil {
		m.PayToPatterns = []string{}
	}
	return nil
}

func (m *PaymentSpendPolicy) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
