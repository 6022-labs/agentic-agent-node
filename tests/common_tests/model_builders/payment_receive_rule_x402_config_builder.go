package model_builders

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/google/uuid"
)

type PaymentReceiveRuleX402ConfigBuilder struct {
	paymentReceiveRuleId uuid.UUID
	network              string
	asset                string
	payTo                string
	assetTransferMethod  enums.AssetTransferMethod
	eip712Name           string
	eip712Version        string
	addressPatterns      []string
}

func NewPaymentReceiveRuleX402ConfigBuilder() *PaymentReceiveRuleX402ConfigBuilder {
	return &PaymentReceiveRuleX402ConfigBuilder{
		paymentReceiveRuleId: uuid.New(),
		network:              "eip155:80002",
		asset:                common_tests.RandomEthAddress().Hex(),
		payTo:                common_tests.RandomEthAddress().Hex(),
		assetTransferMethod:  enums.AssetTransferMethodEip3009,
		eip712Name:           "USDC",
		eip712Version:        "2",
		addressPatterns:      []string{strings.ToLower(common_tests.RandomEthAddress().Hex())},
	}
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithPaymentReceiveRuleId(
	id uuid.UUID,
) *PaymentReceiveRuleX402ConfigBuilder {
	b.paymentReceiveRuleId = id
	return b
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithAddressPatterns(
	patterns []string,
) *PaymentReceiveRuleX402ConfigBuilder {
	b.addressPatterns = patterns
	return b
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithAssetTransferMethod(
	method enums.AssetTransferMethod,
) *PaymentReceiveRuleX402ConfigBuilder {
	b.assetTransferMethod = method
	return b
}

func (b *PaymentReceiveRuleX402ConfigBuilder) Build() *models.PaymentReceiveRuleX402Config {
	config := &models.PaymentReceiveRuleX402Config{
		PaymentReceiveRuleId: b.paymentReceiveRuleId,
		Network:              b.network,
		Asset:                b.asset,
		PayTo:                b.payTo,
		AssetTransferMethod:  b.assetTransferMethod,
		AddressPatterns:      b.addressPatterns,
	}
	// permit2 signs over Permit2's fixed domain, so the token domain stays nil.
	if b.assetTransferMethod == enums.AssetTransferMethodEip3009 {
		config.Eip712Name = &b.eip712Name
		config.Eip712Version = &b.eip712Version
	}
	return config
}
