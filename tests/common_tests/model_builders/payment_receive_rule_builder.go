package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type PaymentReceiveRuleBuilder struct {
	id        uuid.UUID
	policyId  uuid.UUID
	position  int
	name      string
	price     string
	method    enums.PaymentMethod
	createdAt time.Time
	updatedAt time.Time
}

func NewPaymentReceiveRuleBuilder() *PaymentReceiveRuleBuilder {
	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &PaymentReceiveRuleBuilder{
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

func (b *PaymentReceiveRuleBuilder) WithId(id uuid.UUID) *PaymentReceiveRuleBuilder {
	b.id = id
	return b
}

func (b *PaymentReceiveRuleBuilder) WithPolicyId(policyId uuid.UUID) *PaymentReceiveRuleBuilder {
	b.policyId = policyId
	return b
}

func (b *PaymentReceiveRuleBuilder) WithPosition(position int) *PaymentReceiveRuleBuilder {
	b.position = position
	return b
}

func (b *PaymentReceiveRuleBuilder) WithName(name string) *PaymentReceiveRuleBuilder {
	b.name = name
	return b
}

func (b *PaymentReceiveRuleBuilder) WithPrice(price string) *PaymentReceiveRuleBuilder {
	b.price = price
	return b
}

func (b *PaymentReceiveRuleBuilder) WithMethod(method enums.PaymentMethod) *PaymentReceiveRuleBuilder {
	b.method = method
	return b
}

func (b *PaymentReceiveRuleBuilder) Build() *models.PaymentReceiveRule {
	return &models.PaymentReceiveRule{
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
