package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type MCPConnectionBuilder struct {
	id        uuid.UUID
	agentId   uuid.UUID
	name      string
	url       string
	createdAt time.Time
	updatedAt time.Time

	headerAuthKeys      []string
	allowedToolPatterns []string
}

func NewMCPConnectionBuilder() *MCPConnectionBuilder {
	id := uuid.New()
	name := faker.Word()

	url := faker.URL()

	headerAuthKeys := []string{"authorization"}

	createdAtStr := faker.Date()
	updatedAtStr := faker.Date()

	createdAt, _ := time.Parse(time.RFC3339, createdAtStr)
	updatedAt, _ := time.Parse(time.RFC3339, updatedAtStr)

	return &MCPConnectionBuilder{
		id:             id,
		agentId:        uuid.New(),
		url:            url,
		name:           name,
		createdAt:      createdAt,
		updatedAt:      updatedAt,
		headerAuthKeys: headerAuthKeys,
	}
}

func (b *MCPConnectionBuilder) WithId(id uuid.UUID) *MCPConnectionBuilder {
	b.id = id
	return b
}

func (b *MCPConnectionBuilder) WithAgentId(agentId uuid.UUID) *MCPConnectionBuilder {
	b.agentId = agentId
	return b
}

func (b *MCPConnectionBuilder) WithName(name string) *MCPConnectionBuilder {
	b.name = name
	return b
}

func (b *MCPConnectionBuilder) WithUrl(url string) *MCPConnectionBuilder {
	b.url = url
	return b
}

func (b *MCPConnectionBuilder) WithHeaderAuthKeys(headerAuthKeys []string) *MCPConnectionBuilder {
	b.headerAuthKeys = headerAuthKeys
	return b
}

func (b *MCPConnectionBuilder) WithAllowedToolPatterns(allowedToolPatterns []string) *MCPConnectionBuilder {
	b.allowedToolPatterns = allowedToolPatterns
	return b
}

func (b *MCPConnectionBuilder) WithCreatedAt(createdAt time.Time) *MCPConnectionBuilder {
	b.createdAt = createdAt
	return b
}

func (b *MCPConnectionBuilder) WithUpdatedAt(updatedAt time.Time) *MCPConnectionBuilder {
	b.updatedAt = updatedAt
	return b
}

func (b *MCPConnectionBuilder) Build() *models.MCPConnection {
	var headerAuthKeys models.MCPConnectionHeaderAuthKeys
	headerAuthKeys.FromArray(b.headerAuthKeys)

	return &models.MCPConnection{
		Id:                  b.id,
		AgentId:             b.agentId,
		Url:                 b.url,
		Name:                b.name,
		CreatedAt:           b.createdAt,
		UpdatedAt:           b.updatedAt,
		HeaderAuthKeys:      headerAuthKeys,
		AllowedToolPatterns: b.allowedToolPatterns,
	}
}
