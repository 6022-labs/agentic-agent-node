package model_builders

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/samber/lo"
)

type AgentBuilder struct {
	id                uuid.UUID
	chainId           string
	collectionTokenId *uint64
	collectionAddress string
	proposalId        *uint64
	name              string
	cloneOf           *uint64
	ensDomain         string
	creator           string
	owner             *string
	soul              string
	images            map[string]string
	attributes        map[string]string
	createdAt         time.Time
	updatedAt         time.Time
}

func NewAgentBuilder() *AgentBuilder {
	collectionTokenId := uint64(1)
	owner := common_tests.RandomEthAddress().Hex()

	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &AgentBuilder{
		id:                uuid.New(),
		chainId:           "80002",
		collectionTokenId: &collectionTokenId,
		collectionAddress: common_tests.RandomEthAddress().Hex(),
		name:              faker.Name(),
		ensDomain:         faker.Word() + ".agent.eth",
		creator:           common_tests.RandomEthAddress().Hex(),
		owner:             &owner,
		images:            map[string]string{"avatar": faker.URL()},
		attributes:        map[string]string{"role": "facilitator", "trait": faker.Word()},
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (b *AgentBuilder) WithChainId(chainId string) *AgentBuilder {
	b.chainId = chainId
	return b
}

func (b *AgentBuilder) WithId(id uuid.UUID) *AgentBuilder {
	b.id = id
	return b
}

func (b *AgentBuilder) WithCollectionTokenId(tokenId uint64) *AgentBuilder {
	b.collectionTokenId = &tokenId
	return b
}

// WithoutCollectionTokenId models a not-yet-minted draft.
func (b *AgentBuilder) WithoutCollectionTokenId() *AgentBuilder {
	b.collectionTokenId = nil
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

func (b *AgentBuilder) WithCreator(creator string) *AgentBuilder {
	b.creator = creator
	return b
}

func (b *AgentBuilder) WithSoul(soul string) *AgentBuilder {
	b.soul = soul
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
		ChainId:           lo.ToPtr(b.chainId),
		CollectionTokenId: b.collectionTokenId,
		CollectionAddress: b.collectionAddress,
		ProposalId:        b.proposalId,
		Name:              lo.ToPtr(b.name),
		CloneOf:           b.cloneOf,
		EnsDomain:         lo.ToPtr(b.ensDomain),
		Creator:           lo.ToPtr(b.creator),
		Owner:             b.owner,
		Soul:              lo.ToPtr(b.soul),
		Images:            b.images,
		Attributes:        b.attributes,
		CreatedAt:         b.createdAt,
		UpdatedAt:         b.updatedAt,
	}
}
