package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/google/uuid"
)

// AgentAddress is one wallet address row per chain family, keyed by (agent_id, type); the private key lives in Vault.
type AgentAddress struct {
	AgentId   uuid.UUID         `gorm:"primaryKey;size:36" json:"agentId"`
	Type      enums.AddressType `gorm:"primaryKey;size:16" json:"type"`
	Address   string            `gorm:"not null;size:90"   json:"address"`
	UpdatedAt time.Time         `gorm:"not null"           json:"updatedAt"`
	CreatedAt time.Time         `gorm:"not null"           json:"createdAt"`
}

func (AgentAddress) TableName() string {
	return "agent_addresses"
}
