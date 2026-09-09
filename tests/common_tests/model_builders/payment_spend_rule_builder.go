package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
)

type PaymentSpendRuleBuilder struct {
	id                  uuid.UUID
	policyId            uuid.UUID
	name                string
	maxAmountPerPayment string
	windowSeconds       int
	maxAmountPerWindow  string
	windowSpent         string
	windowStartedAt     *time.Time
	createdAt           time.Time
}

// NewPaymentSpendRuleBuilder mirrors the column defaults, so a built rule is one the database could return.
func NewPaymentSpendRuleBuilder() *PaymentSpendRuleBuilder {
	return &PaymentSpendRuleBuilder{
		id:                  uuid.New(),
		policyId:            uuid.New(),
		name:                "Test spend rule",
		maxAmountPerPayment: "1000000",
		maxAmountPerWindow:  "0",
		windowSpent:         "0",
	}
}

func (b *PaymentSpendRuleBuilder) WithId(id uuid.UUID) *PaymentSpendRuleBuilder {
	b.id = id
	return b
}

func (b *PaymentSpendRuleBuilder) WithPolicyId(policyId uuid.UUID) *PaymentSpendRuleBuilder {
	b.policyId = policyId
	return b
}

func (b *PaymentSpendRuleBuilder) WithName(name string) *PaymentSpendRuleBuilder {
	b.name = name
	return b
}

func (b *PaymentSpendRuleBuilder) WithMaxAmountPerPayment(maxAmountPerPayment string) *PaymentSpendRuleBuilder {
	b.maxAmountPerPayment = maxAmountPerPayment
	return b
}

func (b *PaymentSpendRuleBuilder) WithWindowSeconds(windowSeconds int) *PaymentSpendRuleBuilder {
	b.windowSeconds = windowSeconds
	return b
}

func (b *PaymentSpendRuleBuilder) WithMaxAmountPerWindow(maxAmountPerWindow string) *PaymentSpendRuleBuilder {
	b.maxAmountPerWindow = maxAmountPerWindow
	return b
}

func (b *PaymentSpendRuleBuilder) WithWindowSpent(windowSpent string) *PaymentSpendRuleBuilder {
	b.windowSpent = windowSpent
	return b
}

func (b *PaymentSpendRuleBuilder) WithWindowStartedAt(windowStartedAt time.Time) *PaymentSpendRuleBuilder {
	b.windowStartedAt = &windowStartedAt
	return b
}

func (b *PaymentSpendRuleBuilder) WithCreatedAt(createdAt time.Time) *PaymentSpendRuleBuilder {
	b.createdAt = createdAt
	return b
}

func (b *PaymentSpendRuleBuilder) Build() *models.PaymentSpendRule {
	return &models.PaymentSpendRule{
		Id:                  b.id,
		PolicyId:            b.policyId,
		Name:                b.name,
		MaxAmountPerPayment: b.maxAmountPerPayment,
		Method:              enums.PaymentMethodX402,
		WindowSeconds:       b.windowSeconds,
		MaxAmountPerWindow:  b.maxAmountPerWindow,
		WindowSpent:         b.windowSpent,
		WindowStartedAt:     b.windowStartedAt,
		CreatedAt:           b.createdAt,
	}
}
