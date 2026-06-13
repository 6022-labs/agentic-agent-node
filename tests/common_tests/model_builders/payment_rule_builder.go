package model_builders

import (
	"strings"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

type PaymentRuleBuilder struct {
	id              uuid.UUID
	policyId        uuid.UUID
	name            string
	network         string
	asset           string
	payTo           string
	price           string
	addressPatterns []string
	createdAt       time.Time
	updatedAt       time.Time
}

func NewPaymentRuleBuilder() *PaymentRuleBuilder {
	createdAt, _ := time.Parse(time.RFC3339, faker.Date())
	updatedAt, _ := time.Parse(time.RFC3339, faker.Date())

	return &PaymentRuleBuilder{
		id:              uuid.New(),
		policyId:        uuid.New(),
		name:            faker.Word(),
		network:         "eip155:80002",
		asset:           common_tests.RandomEthAddress().Hex(),
		payTo:           common_tests.RandomEthAddress().Hex(),
		price:           "10000",
		addressPatterns: []string{strings.ToLower(common_tests.RandomEthAddress().Hex())},
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (b *PaymentRuleBuilder) WithId(id uuid.UUID) *PaymentRuleBuilder {
	b.id = id
	return b
}

func (b *PaymentRuleBuilder) WithPolicyId(policyId uuid.UUID) *PaymentRuleBuilder {
	b.policyId = policyId
	return b
}

func (b *PaymentRuleBuilder) WithName(name string) *PaymentRuleBuilder {
	b.name = name
	return b
}

func (b *PaymentRuleBuilder) WithNetwork(network string) *PaymentRuleBuilder {
	b.network = network
	return b
}

func (b *PaymentRuleBuilder) WithPrice(price string) *PaymentRuleBuilder {
	b.price = price
	return b
}

func (b *PaymentRuleBuilder) WithAddressPatterns(patterns []string) *PaymentRuleBuilder {
	b.addressPatterns = patterns
	return b
}

func (b *PaymentRuleBuilder) Build() *models.PaymentRule {
	return &models.PaymentRule{
		Id:                b.id,
		PolicyId:          b.policyId,
		Name:              b.name,
		Network:           b.network,
		Asset:             b.asset,
		PayTo:             b.payTo,
		Price:             b.price,
		MaxTimeoutSeconds: 60,
		Eip712Name:        "USDC",
		Eip712Version:     "2",
		AddressPatterns:   b.addressPatterns,
		CreatedAt:         b.createdAt,
		UpdatedAt:         b.updatedAt,
	}
}
