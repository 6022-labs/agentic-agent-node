package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type ConversationMessageBuilder struct {
	id             uuid.UUID
	conversationId string
	author         string
	role           string
	platform       string
	content        string
	hash           string
	timestamp      time.Time
	createdAt      time.Time
}

func NewConversationMessageBuilder() *ConversationMessageBuilder {
	return &ConversationMessageBuilder{
		id:             uuid.New(),
		conversationId: faker.UUIDHyphenated(),
		author:         faker.Username(),
		role:           "user",
		platform:       "slack",
		content:        faker.Sentence(),
		timestamp:      time.Now().UTC().Truncate(time.Microsecond),
	}
}

func (b *ConversationMessageBuilder) WithId(id uuid.UUID) *ConversationMessageBuilder {
	b.id = id
	return b
}

func (b *ConversationMessageBuilder) WithConversationId(conversationId string) *ConversationMessageBuilder {
	b.conversationId = conversationId
	return b
}

func (b *ConversationMessageBuilder) WithAuthor(author string) *ConversationMessageBuilder {
	b.author = author
	return b
}

func (b *ConversationMessageBuilder) WithRole(role string) *ConversationMessageBuilder {
	b.role = role
	return b
}

func (b *ConversationMessageBuilder) WithPlatform(platform string) *ConversationMessageBuilder {
	b.platform = platform
	return b
}

func (b *ConversationMessageBuilder) WithContent(content string) *ConversationMessageBuilder {
	b.content = content
	return b
}

func (b *ConversationMessageBuilder) WithHash(hash string) *ConversationMessageBuilder {
	b.hash = hash
	return b
}

func (b *ConversationMessageBuilder) WithTimestamp(timestamp time.Time) *ConversationMessageBuilder {
	b.timestamp = timestamp
	return b
}

func (b *ConversationMessageBuilder) WithCreatedAt(createdAt time.Time) *ConversationMessageBuilder {
	b.createdAt = createdAt
	return b
}

func (b *ConversationMessageBuilder) Build() models.ConversationMessage {
	msg := models.ConversationMessage{
		Id:             b.id,
		ConversationId: b.conversationId,
		Author:         b.author,
		Role:           b.role,
		Platform:       b.platform,
		Content:        b.content,
		Hash:           b.hash,
		Timestamp:      b.timestamp,
		CreatedAt:      b.createdAt,
	}
	if msg.Hash == "" {
		_ = msg.ComputeHash()
	}
	return msg
}
