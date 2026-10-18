package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type ConversationBuilder struct {
	id        string
	name      *string
	callerId  *string
	agentId   uuid.UUID
	createdAt time.Time
	updatedAt time.Time
}

func NewConversationBuilder() *ConversationBuilder {
	return &ConversationBuilder{
		id:        faker.UUIDHyphenated(),
		agentId:   uuid.New(),
		createdAt: time.Now().UTC().Truncate(time.Microsecond),
		updatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
}

func (b *ConversationBuilder) WithId(id string) *ConversationBuilder {
	b.id = id
	return b
}

func (b *ConversationBuilder) WithName(name string) *ConversationBuilder {
	b.name = &name
	return b
}

func (b *ConversationBuilder) WithCallerId(callerId string) *ConversationBuilder {
	b.callerId = &callerId
	return b
}

func (b *ConversationBuilder) WithAgentId(agentId uuid.UUID) *ConversationBuilder {
	b.agentId = agentId
	return b
}

func (b *ConversationBuilder) WithCreatedAt(createdAt time.Time) *ConversationBuilder {
	b.createdAt = createdAt
	return b
}

func (b *ConversationBuilder) WithUpdatedAt(updatedAt time.Time) *ConversationBuilder {
	b.updatedAt = updatedAt
	return b
}

func (b *ConversationBuilder) Build() *models.Conversation {
	return &models.Conversation{
		Id:        b.id,
		Name:      b.name,
		CallerId:  b.callerId,
		AgentId:   b.agentId,
		CreatedAt: b.createdAt,
		UpdatedAt: b.updatedAt,
	}
}
