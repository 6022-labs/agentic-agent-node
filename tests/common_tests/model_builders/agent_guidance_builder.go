package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type AgentGuidanceBuilder struct {
	id        uuid.UUID
	agentId   uuid.UUID
	content   string
	createdAt time.Time
	updatedAt time.Time
}

func NewAgentGuidanceBuilder() *AgentGuidanceBuilder {
	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &AgentGuidanceBuilder{
		id:        uuid.New(),
		agentId:   uuid.New(),
		content:   faker.Sentence(),
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (b *AgentGuidanceBuilder) WithId(id uuid.UUID) *AgentGuidanceBuilder {
	b.id = id
	return b
}

func (b *AgentGuidanceBuilder) WithAgentId(agentId uuid.UUID) *AgentGuidanceBuilder {
	b.agentId = agentId
	return b
}

func (b *AgentGuidanceBuilder) WithContent(content string) *AgentGuidanceBuilder {
	b.content = content
	return b
}

func (b *AgentGuidanceBuilder) WithCreatedAt(createdAt time.Time) *AgentGuidanceBuilder {
	b.createdAt = createdAt
	return b
}

func (b *AgentGuidanceBuilder) WithUpdatedAt(updatedAt time.Time) *AgentGuidanceBuilder {
	b.updatedAt = updatedAt
	return b
}

func (b *AgentGuidanceBuilder) Build() *models.AgentGuidance {
	return &models.AgentGuidance{
		Id:        b.id,
		AgentId:   b.agentId,
		Content:   b.content,
		CreatedAt: b.createdAt,
		UpdatedAt: b.updatedAt,
	}
}
