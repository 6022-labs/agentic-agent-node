package model_builders

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
)

type PaymentSpendRuleBuilder struct {
	id        uuid.UUID
	policyId  uuid.UUID
	position  int
	name      string
	maxAmount string
	method    enums.PaymentMethod
}

func NewPaymentSpendRuleBuilder() *PaymentSpendRuleBuilder {
	return &PaymentSpendRuleBuilder{
		id:        uuid.New(),
		policyId:  uuid.New(),
		name:      "Test spend rule",
		maxAmount: "1000000",
		method:    enums.PaymentMethodX402,
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

func (b *PaymentSpendRuleBuilder) WithPosition(position int) *PaymentSpendRuleBuilder {
	b.position = position
	return b
}

func (b *PaymentSpendRuleBuilder) WithName(name string) *PaymentSpendRuleBuilder {
	b.name = name
	return b
}

func (b *PaymentSpendRuleBuilder) WithMaxAmount(maxAmount string) *PaymentSpendRuleBuilder {
	b.maxAmount = maxAmount
	return b
}

func (b *PaymentSpendRuleBuilder) WithMethod(method enums.PaymentMethod) *PaymentSpendRuleBuilder {
	b.method = method
	return b
}

func (b *PaymentSpendRuleBuilder) Build() *models.PaymentSpendRule {
	return &models.PaymentSpendRule{
		Id:        b.id,
		PolicyId:  b.policyId,
		Position:  b.position,
		Name:      b.name,
		MaxAmount: b.maxAmount,
		Method:    b.method,
	}
}
