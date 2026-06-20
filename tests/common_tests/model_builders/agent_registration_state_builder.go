package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/google/uuid"
)

type AgentRegistrationStateBuilder struct {
	id                uuid.UUID
	agentId           uuid.UUID
	mintStatus        enums.TransactionStatus
	mintChainId       *uint64
	mintTxHash        *string
	mintedAt          *time.Time
	ensRegistryStatus enums.TransactionStatus
	ensChainId        *uint64
	ensTxHash         *string
	ensPublishedAt    *time.Time
	createdAt         time.Time
	updatedAt         time.Time
}

func NewAgentRegistrationStateBuilder() *AgentRegistrationStateBuilder {
	createdAt, _ := time.Parse(time.RFC3339, "2026-01-01T00:00:00Z")
	updatedAt, _ := time.Parse(time.RFC3339, "2026-01-02T00:00:00Z")

	return &AgentRegistrationStateBuilder{
		id:                uuid.New(),
		agentId:           uuid.New(),
		mintStatus:        enums.TransactionStatusPending,
		ensRegistryStatus: enums.TransactionStatusPending,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (b *AgentRegistrationStateBuilder) WithId(id uuid.UUID) *AgentRegistrationStateBuilder {
	b.id = id
	return b
}

func (b *AgentRegistrationStateBuilder) WithAgentId(agentId uuid.UUID) *AgentRegistrationStateBuilder {
	b.agentId = agentId
	return b
}

func (b *AgentRegistrationStateBuilder) WithMintStatus(status enums.TransactionStatus) *AgentRegistrationStateBuilder {
	b.mintStatus = status
	return b
}

func (b *AgentRegistrationStateBuilder) WithMintChainId(chainId uint64) *AgentRegistrationStateBuilder {
	b.mintChainId = &chainId
	return b
}

func (b *AgentRegistrationStateBuilder) WithMintTxHash(txHash string) *AgentRegistrationStateBuilder {
	b.mintTxHash = &txHash
	return b
}

func (b *AgentRegistrationStateBuilder) WithMintedAt(mintedAt time.Time) *AgentRegistrationStateBuilder {
	b.mintedAt = &mintedAt
	return b
}

func (b *AgentRegistrationStateBuilder) WithEnsRegistryStatus(status enums.TransactionStatus) *AgentRegistrationStateBuilder {
	b.ensRegistryStatus = status
	return b
}

func (b *AgentRegistrationStateBuilder) WithEnsChainId(chainId uint64) *AgentRegistrationStateBuilder {
	b.ensChainId = &chainId
	return b
}

func (b *AgentRegistrationStateBuilder) WithEnsTxHash(txHash string) *AgentRegistrationStateBuilder {
	b.ensTxHash = &txHash
	return b
}

func (b *AgentRegistrationStateBuilder) WithEnsPublishedAt(publishedAt time.Time) *AgentRegistrationStateBuilder {
	b.ensPublishedAt = &publishedAt
	return b
}

func (b *AgentRegistrationStateBuilder) WithCreatedAt(createdAt time.Time) *AgentRegistrationStateBuilder {
	b.createdAt = createdAt
	return b
}

func (b *AgentRegistrationStateBuilder) WithUpdatedAt(updatedAt time.Time) *AgentRegistrationStateBuilder {
	b.updatedAt = updatedAt
	return b
}

func (b *AgentRegistrationStateBuilder) Build() *models.AgentRegistrationState {
	return &models.AgentRegistrationState{
		Id:                b.id,
		AgentId:           b.agentId,
		MintStatus:        b.mintStatus,
		MintChainId:       b.mintChainId,
		MintTxHash:        b.mintTxHash,
		MintedAt:          b.mintedAt,
		EnsRegistryStatus: b.ensRegistryStatus,
		EnsChainId:        b.ensChainId,
		EnsTxHash:         b.ensTxHash,
		EnsPublishedAt:    b.ensPublishedAt,
		CreatedAt:         b.createdAt,
		UpdatedAt:         b.updatedAt,
	}
}
