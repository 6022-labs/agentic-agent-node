package model_builders

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/google/uuid"
)

type PaymentRuleX402ConfigBuilder struct {
	paymentRuleId   uuid.UUID
	network         string
	asset           string
	payTo           string
	eip712Name      string
	eip712Version   string
	addressPatterns []string
}

func NewPaymentRuleX402ConfigBuilder() *PaymentRuleX402ConfigBuilder {
	return &PaymentRuleX402ConfigBuilder{
		paymentRuleId:   uuid.New(),
		network:         "eip155:80002",
		asset:           common_tests.RandomEthAddress().Hex(),
		payTo:           common_tests.RandomEthAddress().Hex(),
		eip712Name:      "USDC",
		eip712Version:   "2",
		addressPatterns: []string{strings.ToLower(common_tests.RandomEthAddress().Hex())},
	}
}

func (b *PaymentRuleX402ConfigBuilder) WithPaymentRuleId(id uuid.UUID) *PaymentRuleX402ConfigBuilder {
	b.paymentRuleId = id
	return b
}

func (b *PaymentRuleX402ConfigBuilder) WithAddressPatterns(patterns []string) *PaymentRuleX402ConfigBuilder {
	b.addressPatterns = patterns
	return b
}

func (b *PaymentRuleX402ConfigBuilder) Build() *models.PaymentRuleX402Config {
	return &models.PaymentRuleX402Config{
		PaymentRuleId:   b.paymentRuleId,
		Network:         b.network,
		Asset:           b.asset,
		PayTo:           b.payTo,
		Eip712Name:      b.eip712Name,
		Eip712Version:   b.eip712Version,
		AddressPatterns: b.addressPatterns,
	}
}
