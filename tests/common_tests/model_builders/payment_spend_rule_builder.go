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
	maxAmountPerPayment *string
	createdAt           time.Time
}

func NewPaymentSpendRuleBuilder() *PaymentSpendRuleBuilder {
	maxAmountPerPayment := "1000000"
	return &PaymentSpendRuleBuilder{
		id:                  uuid.New(),
		policyId:            uuid.New(),
		name:                "Test spend rule",
		maxAmountPerPayment: &maxAmountPerPayment,
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
	b.maxAmountPerPayment = &maxAmountPerPayment
	return b
}

func (b *PaymentSpendRuleBuilder) WithoutMaxAmountPerPayment() *PaymentSpendRuleBuilder {
	b.maxAmountPerPayment = nil
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
		Method:              enums.PaymentMethodX402,
		MaxAmountPerPayment: b.maxAmountPerPayment,
		CreatedAt:           b.createdAt,
	}
}
