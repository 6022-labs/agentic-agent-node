package model_builders

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

type PaymentSpendPolicyBuilder struct {
	id            uuid.UUID
	agentId       uuid.UUID
	payToPatterns []string
}

// NewPaymentSpendPolicyBuilder allows paying anyone, so a test denies a payee only when it says so.
func NewPaymentSpendPolicyBuilder() *PaymentSpendPolicyBuilder {
	return &PaymentSpendPolicyBuilder{
		id:            uuid.New(),
		agentId:       uuid.New(),
		payToPatterns: []string{".*"},
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

func (b *PaymentSpendPolicyBuilder) WithPayToPatterns(payToPatterns []string) *PaymentSpendPolicyBuilder {
	b.payToPatterns = payToPatterns
	return b
}

func (b *PaymentSpendPolicyBuilder) Build() *models.PaymentSpendPolicy {
	return &models.PaymentSpendPolicy{
		Id:            b.id,
		AgentId:       b.agentId,
		PayToPatterns: b.payToPatterns,
	}
}
