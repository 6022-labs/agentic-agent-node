package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type PaymentPolicyBuilder struct {
	id        uuid.UUID
	agentId   uuid.UUID
	createdAt time.Time
	updatedAt time.Time
}

func NewPaymentPolicyBuilder() *PaymentPolicyBuilder {
	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &PaymentPolicyBuilder{
		id:        uuid.New(),
		agentId:   uuid.New(),
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (b *PaymentPolicyBuilder) WithId(id uuid.UUID) *PaymentPolicyBuilder {
	b.id = id
	return b
}

func (b *PaymentPolicyBuilder) WithAgentId(agentId uuid.UUID) *PaymentPolicyBuilder {
	b.agentId = agentId
	return b
}

func (b *PaymentPolicyBuilder) Build() *models.PaymentPolicy {
	return &models.PaymentPolicy{
		Id:        b.id,
		AgentId:   b.agentId,
		CreatedAt: b.createdAt,
		UpdatedAt: b.updatedAt,
	}
}
