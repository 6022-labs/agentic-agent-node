package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type PaymentRuleBuilder struct {
	id        uuid.UUID
	policyId  uuid.UUID
	position  int
	name      string
	price     string
	method    enums.PaymentMethod
	createdAt time.Time
	updatedAt time.Time
}

func NewPaymentRuleBuilder() *PaymentRuleBuilder {
	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &PaymentRuleBuilder{
		id:        uuid.New(),
		policyId:  uuid.New(),
		position:  0,
		name:      faker.Word(),
		price:     "10000",
		method:    enums.PaymentMethodX402,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (b *PaymentRuleBuilder) WithId(id uuid.UUID) *PaymentRuleBuilder {
	b.id = id
	return b
}

func (b *PaymentRuleBuilder) WithPolicyId(policyId uuid.UUID) *PaymentRuleBuilder {
	b.policyId = policyId
	return b
}

func (b *PaymentRuleBuilder) WithPosition(position int) *PaymentRuleBuilder {
	b.position = position
	return b
}

func (b *PaymentRuleBuilder) WithName(name string) *PaymentRuleBuilder {
	b.name = name
	return b
}

func (b *PaymentRuleBuilder) WithPrice(price string) *PaymentRuleBuilder {
	b.price = price
	return b
}

func (b *PaymentRuleBuilder) WithMethod(method enums.PaymentMethod) *PaymentRuleBuilder {
	b.method = method
	return b
}

func (b *PaymentRuleBuilder) Build() *models.PaymentRule {
	return &models.PaymentRule{
		Id:        b.id,
		PolicyId:  b.policyId,
		Position:  b.position,
		Name:      b.name,
		Price:     b.price,
		Method:    b.method,
		CreatedAt: b.createdAt,
		UpdatedAt: b.updatedAt,
	}
}
