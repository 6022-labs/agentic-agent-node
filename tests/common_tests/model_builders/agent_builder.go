package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type AgentBuilder struct {
	id                uuid.UUID
	collectionTokenId *uint64
	collectionAddress string
	proposalId        *uint64
	name              string
	cloneOf           *uint64
	ensDomain         string
	role              enums.AgentRole
	creator           string
	walletAddress     string
	owner             *string
	images            map[string]string
	attributes        map[string]string
	createdAt         time.Time
	updatedAt         time.Time
}

func NewAgentBuilder() *AgentBuilder {
	collectionTokenId := uint64(1)
	owner := common_tests.RandomEthAddress().Hex()

	roles := enums.GetAllAgentRoles()
	roleIndex, _ := faker.RandomInt(0, len(roles)-1, 1)
	role := roles[roleIndex[0]]

	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &AgentBuilder{
		id:                uuid.New(),
		collectionTokenId: &collectionTokenId,
		collectionAddress: common_tests.RandomEthAddress().Hex(),
		name:              faker.Name(),
		ensDomain:         faker.Word() + ".agent.eth",
		role:              role,
		creator:           common_tests.RandomEthAddress().Hex(),
		walletAddress:     common_tests.RandomEthAddress().Hex(),
		owner:             &owner,
		images:            map[string]string{"avatar": faker.URL()},
		attributes:        map[string]string{"trait": faker.Word()},
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (b *AgentBuilder) WithId(id uuid.UUID) *AgentBuilder {
	b.id = id
	return b
}

func (b *AgentBuilder) WithCollectionTokenId(tokenId uint64) *AgentBuilder {
	b.collectionTokenId = &tokenId
	return b
}

func (b *AgentBuilder) WithCollectionAddress(address string) *AgentBuilder {
	b.collectionAddress = address
	return b
}

func (b *AgentBuilder) WithProposalId(proposalId uint64) *AgentBuilder {
	b.proposalId = &proposalId
	return b
}

func (b *AgentBuilder) WithName(name string) *AgentBuilder {
	b.name = name
	return b
}

func (b *AgentBuilder) WithCloneOf(cloneOf uint64) *AgentBuilder {
	b.cloneOf = &cloneOf
	return b
}

func (b *AgentBuilder) WithEnsDomain(ensDomain string) *AgentBuilder {
	b.ensDomain = ensDomain
	return b
}

func (b *AgentBuilder) WithRole(role enums.AgentRole) *AgentBuilder {
	b.role = role
	return b
}

func (b *AgentBuilder) WithCreator(creator string) *AgentBuilder {
	b.creator = creator
	return b
}

func (b *AgentBuilder) WithWalletAddress(walletAddress string) *AgentBuilder {
	b.walletAddress = walletAddress
	return b
}

func (b *AgentBuilder) WithOwner(owner string) *AgentBuilder {
	b.owner = &owner
	return b
}

func (b *AgentBuilder) WithImages(images map[string]string) *AgentBuilder {
	b.images = images
	return b
}

func (b *AgentBuilder) WithAttributes(attributes map[string]string) *AgentBuilder {
	b.attributes = attributes
	return b
}

func (b *AgentBuilder) WithCreatedAt(createdAt time.Time) *AgentBuilder {
	b.createdAt = createdAt
	return b
}

func (b *AgentBuilder) WithUpdatedAt(updatedAt time.Time) *AgentBuilder {
	b.updatedAt = updatedAt
	return b
}

func (b *AgentBuilder) Build() *models.Agent {
	return &models.Agent{
		Id:                b.id,
		CollectionTokenId: b.collectionTokenId,
		CollectionAddress: b.collectionAddress,
		ProposalId:        b.proposalId,
		Name:              b.name,
		CloneOf:           b.cloneOf,
		EnsDomain:         b.ensDomain,
		Role:              b.role,
		Creator:           b.creator,
		WalletAddress:     b.walletAddress,
		Owner:             b.owner,
		Images:            b.images,
		Attributes:        b.attributes,
		CreatedAt:         b.createdAt,
		UpdatedAt:         b.updatedAt,
	}
}
