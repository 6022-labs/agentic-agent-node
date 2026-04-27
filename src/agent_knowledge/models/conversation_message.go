package models

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversationMessage is one message the agent has seen in a given conversation.
// Schema is owned by the goose migration in agent_knowledge_db/migrations/.
//
// Dedup is by `Hash` (SHA-512 over role/author/content/platform/conversationId)
// — mirrors the orchestrator's database_connector contract. `Id` is a locally
// generated uuid for traceability.
type ConversationMessage struct {
	Id             uuid.UUID `gorm:"primaryKey"`
	ConversationId string
	Author         string
	Role           string
	Platform       string
	Content        string
	Hash           string
	Timestamp      time.Time
	CreatedAt      time.Time
}

func (ConversationMessage) TableName() string {
	return "conversation_messages"
}

func (m *ConversationMessage) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.Hash == "" {
		return m.ComputeHash()
	}
	return nil
}

// ComputeHash sets Hash from the semantic identity of the message — the
// fields that make two messages "the same" regardless of when or by whom they
// were stored.
func (m *ConversationMessage) ComputeHash() error {
	bytes, err := json.Marshal(struct {
		Role           string `json:"role"`
		Author         string `json:"author"`
		Content        string `json:"content"`
		Platform       string `json:"platform"`
		ConversationId string `json:"conversationId"`
	}{
		Role:           m.Role,
		Author:         m.Author,
		Content:        m.Content,
		Platform:       m.Platform,
		ConversationId: m.ConversationId,
	})
	if err != nil {
		return err
	}
	sum := sha512.Sum512(bytes)
	m.Hash = hex.EncodeToString(sum[:])
	return nil
}
