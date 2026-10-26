package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type PaymentReceivePolicyBuilder struct {
	id        uuid.UUID
	agentId   uuid.UUID
	createdAt time.Time
	updatedAt time.Time
}

func NewPaymentReceivePolicyBuilder() *PaymentReceivePolicyBuilder {
	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &PaymentReceivePolicyBuilder{
		id:        uuid.New(),
		agentId:   uuid.New(),
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (b *PaymentReceivePolicyBuilder) WithId(id uuid.UUID) *PaymentReceivePolicyBuilder {
	b.id = id
	return b
}

func (b *PaymentReceivePolicyBuilder) WithAgentId(agentId uuid.UUID) *PaymentReceivePolicyBuilder {
	b.agentId = agentId
	return b
}

func (b *PaymentReceivePolicyBuilder) Build() *models.PaymentReceivePolicy {
	return &models.PaymentReceivePolicy{
		Id:        b.id,
		AgentId:   b.agentId,
		CreatedAt: b.createdAt,
		UpdatedAt: b.updatedAt,
	}
}
