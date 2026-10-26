package model_builders

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

type PaymentSpendPolicyBuilder struct {
	id      uuid.UUID
	agentId uuid.UUID
}

func NewPaymentSpendPolicyBuilder() *PaymentSpendPolicyBuilder {
	return &PaymentSpendPolicyBuilder{
		id:      uuid.New(),
		agentId: uuid.New(),
	}
}

func (b *PaymentSpendPolicyBuilder) WithId(id uuid.UUID) *PaymentSpendPolicyBuilder {
	b.id = id
	return b
}

func (b *PaymentSpendPolicyBuilder) WithAgentId(agentId uuid.UUID) *PaymentSpendPolicyBuilder {
	b.agentId = agentId
	return b
}

func (b *PaymentSpendPolicyBuilder) Build() *models.PaymentSpendPolicy {
	return &models.PaymentSpendPolicy{
		Id:      b.id,
		AgentId: b.agentId,
	}
}
