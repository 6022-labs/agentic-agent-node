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
	policyId             uuid.UUID
	network              string
	asset                string
	payTo                string
	assetTransferMethod  enums.AssetTransferMethod
	eip712Name           string
	eip712Version        string
}

func NewPaymentReceiveRuleX402ConfigBuilder() *PaymentReceiveRuleX402ConfigBuilder {
	return &PaymentReceiveRuleX402ConfigBuilder{
		paymentReceiveRuleId: uuid.New(),
		policyId:             uuid.New(),
		network:              "eip155:80002",
		asset:                strings.ToLower(common_tests.RandomEthAddress().Hex()),
		payTo:                common_tests.RandomEthAddress().Hex(),
		assetTransferMethod:  enums.AssetTransferMethodEip3009,
		eip712Name:           "USDC",
		eip712Version:        "2",
	}
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithPaymentReceiveRuleId(
	id uuid.UUID,
) *PaymentReceiveRuleX402ConfigBuilder {
	b.paymentReceiveRuleId = id
	return b
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithPolicyId(policyId uuid.UUID) *PaymentReceiveRuleX402ConfigBuilder {
	b.policyId = policyId
	return b
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithNetwork(network string) *PaymentReceiveRuleX402ConfigBuilder {
	b.network = network
	return b
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithAsset(asset string) *PaymentReceiveRuleX402ConfigBuilder {
	b.asset = asset
	return b
}

func (b *PaymentReceiveRuleX402ConfigBuilder) WithPayTo(payTo string) *PaymentReceiveRuleX402ConfigBuilder {
	b.payTo = payTo
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
		PolicyId:             b.policyId,
		Network:              b.network,
		Asset:                b.asset,
		PayTo:                b.payTo,
		AssetTransferMethod:  b.assetTransferMethod,
	}
	// permit2 signs over Permit2's fixed domain, so the token domain stays nil.
	if b.assetTransferMethod == enums.AssetTransferMethodEip3009 {
		config.Eip712Name = &b.eip712Name
		config.Eip712Version = &b.eip712Version
	}
	return config
}
