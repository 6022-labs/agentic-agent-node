package model_builders

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/google/uuid"
)

type PaymentReceiveRulePriceBuilder struct {
	id                   uuid.UUID
	paymentReceiveRuleId uuid.UUID
	price                string
	addressPatterns      []string
}

func NewPaymentReceiveRulePriceBuilder() *PaymentReceiveRulePriceBuilder {
	return &PaymentReceiveRulePriceBuilder{
		id:                   uuid.New(),
		paymentReceiveRuleId: uuid.New(),
		price:                "10000",
		addressPatterns:      []string{strings.ToLower(common_tests.RandomEthAddress().Hex())},
	}
}

func (b *PaymentReceiveRulePriceBuilder) WithId(id uuid.UUID) *PaymentReceiveRulePriceBuilder {
	b.id = id
	return b
}

func (b *PaymentReceiveRulePriceBuilder) WithPaymentReceiveRuleId(id uuid.UUID) *PaymentReceiveRulePriceBuilder {
	b.paymentReceiveRuleId = id
	return b
}

func (b *PaymentReceiveRulePriceBuilder) WithPrice(price string) *PaymentReceiveRulePriceBuilder {
	b.price = price
	return b
}

func (b *PaymentReceiveRulePriceBuilder) WithAddressPatterns(patterns []string) *PaymentReceiveRulePriceBuilder {
	b.addressPatterns = patterns
	return b
}

func (b *PaymentReceiveRulePriceBuilder) Build() *models.PaymentReceiveRulePrice {
	return &models.PaymentReceiveRulePrice{
		Id:                   b.id,
		PaymentReceiveRuleId: b.paymentReceiveRuleId,
		Price:                b.price,
		AddressPatterns:      b.addressPatterns,
	}
}
