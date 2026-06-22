package models

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentRegistrationState tracks the agent's mint (Amoy) and ENS (Ethereum)
// lifecycles, which advance independently. At most one row.
type AgentRegistrationState struct {
	Id      uuid.UUID `gorm:"primaryKey;size:36"`
	AgentId uuid.UUID `gorm:"not null;size:36;uniqueIndex"`
	Agent   Agent     `gorm:"foreignKey:AgentId;references:Id;constraint:OnDelete:CASCADE"`

	MintStatus  enums.TransactionStatus `gorm:"not null;size:32"`
	MintChainId *uint64
	MintTxHash  *string `gorm:"size:66"`
	MintedAt    *time.Time

	EnsRegistryStatus enums.TransactionStatus `gorm:"not null;size:32"`
	EnsChainId        *uint64
	EnsTxHash         *string `gorm:"size:66"`
	EnsPublishedAt    *time.Time

	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (AgentRegistrationState) TableName() string {
	return "agent_registration_states"
}

// IsMinted reports whether the agent's NFT mint is confirmed on chain. Nil-safe.
func (s *AgentRegistrationState) IsMinted() bool {
	return s != nil && s.MintStatus == enums.TransactionStatusConfirmed
}

// MarkMintConfirmed flips the mint lifecycle to confirmed and stamps the time.
func (s *AgentRegistrationState) MarkMintConfirmed() {
	now := time.Now()
	s.MintStatus = enums.TransactionStatusConfirmed
	s.MintedAt = &now
}

// MarkEnsConfirmed flips the ENS lifecycle to confirmed and stamps the time.
func (s *AgentRegistrationState) MarkEnsConfirmed() {
	now := time.Now()
	s.EnsRegistryStatus = enums.TransactionStatusConfirmed
	s.EnsPublishedAt = &now
}

func (m *AgentRegistrationState) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *AgentRegistrationState) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
