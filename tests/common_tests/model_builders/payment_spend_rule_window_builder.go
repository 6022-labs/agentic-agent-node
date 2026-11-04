package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

type PaymentSpendRuleWindowBuilder struct {
	id                 uuid.UUID
	paymentSpendRuleId uuid.UUID
	windowSeconds      int
	maxAmount          string
	spent              string
	startedAt          *time.Time
}

// NewPaymentSpendRuleWindowBuilder is a day-long window that has not started yet.
func NewPaymentSpendRuleWindowBuilder() *PaymentSpendRuleWindowBuilder {
	return &PaymentSpendRuleWindowBuilder{
		id:                 uuid.New(),
		paymentSpendRuleId: uuid.New(),
		windowSeconds:      86400,
		maxAmount:          "10000000",
		spent:              "0",
	}
}

func (b *PaymentSpendRuleWindowBuilder) WithId(id uuid.UUID) *PaymentSpendRuleWindowBuilder {
	b.id = id
	return b
}

func (b *PaymentSpendRuleWindowBuilder) WithPaymentSpendRuleId(id uuid.UUID) *PaymentSpendRuleWindowBuilder {
	b.paymentSpendRuleId = id
	return b
}

func (b *PaymentSpendRuleWindowBuilder) WithWindowSeconds(windowSeconds int) *PaymentSpendRuleWindowBuilder {
	b.windowSeconds = windowSeconds
	return b
}

func (b *PaymentSpendRuleWindowBuilder) WithMaxAmount(maxAmount string) *PaymentSpendRuleWindowBuilder {
	b.maxAmount = maxAmount
	return b
}

func (b *PaymentSpendRuleWindowBuilder) WithSpent(spent string) *PaymentSpendRuleWindowBuilder {
	b.spent = spent
	return b
}

func (b *PaymentSpendRuleWindowBuilder) WithStartedAt(startedAt time.Time) *PaymentSpendRuleWindowBuilder {
	b.startedAt = &startedAt
	return b
}

func (b *PaymentSpendRuleWindowBuilder) Build() *models.PaymentSpendRuleWindow {
	return &models.PaymentSpendRuleWindow{
		Id:                 b.id,
		PaymentSpendRuleId: b.paymentSpendRuleId,
		WindowSeconds:      b.windowSeconds,
		MaxAmount:          b.maxAmount,
		Spent:              b.spent,
		StartedAt:          b.startedAt,
	}
}
