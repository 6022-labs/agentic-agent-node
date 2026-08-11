package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Agent is the agent's persisted on-chain identity. One agent per container, so
// the table holds at most one row. Mint/ENS progress lives in AgentRegistrationState.
type Agent struct {
	Id uuid.UUID `gorm:"primaryKey;size:36" json:"id"`

	ChainId *string `gorm:"size:64" json:"chainId"`

	CollectionTokenId *uint64 `json:"collectionTokenId"`
	CollectionAddress string  `json:"collectionAddress" gorm:"size:42"`
	ProposalId        *uint64 `json:"proposalId"`

	Name      *string `gorm:"size:255" json:"name"`
	CloneOf   *uint64 `                json:"cloneOf"`
	EnsDomain *string `gorm:"size:255" json:"ensDomain"`
	Creator   *string `gorm:"size:42"  json:"creator"`
	Owner     *string `gorm:"size:42"  json:"owner"`

	// Soul is the persona layer of the system prompt, recomputed from on-chain facts at each mint reconcile.
	Soul *string `gorm:"type:text" json:"soul"`

	Images     map[string]string `gorm:"serializer:json" json:"images"`
	Attributes map[string]string `gorm:"serializer:json" json:"attributes"`

	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
}

func (Agent) TableName() string {
	return "agents"
}

func (m *Agent) BeforeCreate(_ *gorm.DB) error {
	if m.Id == uuid.Nil {
		m.Id = uuid.New()
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	return nil
}

func (m *Agent) BeforeSave(_ *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}
