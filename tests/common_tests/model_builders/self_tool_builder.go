package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type SelfToolBuilder struct {
	id        uuid.UUID
	agentId   uuid.UUID
	name      string
	enabled   bool
	createdAt time.Time
	updatedAt time.Time
}

func NewSelfToolBuilder() *SelfToolBuilder {
	createdAtStr := faker.Date()
	updatedAtStr := faker.Date()

	createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
	updatedAt, _ := time.Parse(time.RFC3339, updatedAtStr)

	return &SelfToolBuilder{
		id:        uuid.New(),
		agentId:   uuid.New(),
		name:      faker.Word(),
		enabled:   true,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (b *SelfToolBuilder) WithId(id uuid.UUID) *SelfToolBuilder {
	b.id = id
	return b
}

func (b *SelfToolBuilder) WithAgentId(agentId uuid.UUID) *SelfToolBuilder {
	b.agentId = agentId
	return b
}

func (b *SelfToolBuilder) WithName(name string) *SelfToolBuilder {
	b.name = name
	return b
}

func (b *SelfToolBuilder) WithEnabled(enabled bool) *SelfToolBuilder {
	b.enabled = enabled
	return b
}

func (b *SelfToolBuilder) WithCreatedAt(createdAt time.Time) *SelfToolBuilder {
	b.createdAt = createdAt
	return b
}

func (b *SelfToolBuilder) WithUpdatedAt(updatedAt time.Time) *SelfToolBuilder {
	b.updatedAt = updatedAt
	return b
}

func (b *SelfToolBuilder) Build() *models.SelfTool {
	return &models.SelfTool{
		Id:        b.id,
		AgentId:   b.agentId,
		Name:      b.name,
		Enabled:   b.enabled,
		CreatedAt: b.createdAt,
		UpdatedAt: b.updatedAt,
	}
}
