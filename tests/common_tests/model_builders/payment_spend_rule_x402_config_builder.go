package model_builders

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/google/uuid"
)

type PaymentSpendRuleX402ConfigBuilder struct {
	paymentSpendRuleId  uuid.UUID
	policyId            uuid.UUID
	network             string
	asset               string
	assetTransferMethod *enums.AssetTransferMethod
}

// NewPaymentSpendRuleX402ConfigBuilder names one token on one chain, moved either way.
func NewPaymentSpendRuleX402ConfigBuilder() *PaymentSpendRuleX402ConfigBuilder {
	return &PaymentSpendRuleX402ConfigBuilder{
		paymentSpendRuleId: uuid.New(),
		policyId:           uuid.New(),
		network:            "eip155:80002",
		asset:              strings.ToLower(common_tests.RandomEthAddress().Hex()),
	}
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithPaymentSpendRuleId(id uuid.UUID) *PaymentSpendRuleX402ConfigBuilder {
	b.paymentSpendRuleId = id
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithNetwork(network string) *PaymentSpendRuleX402ConfigBuilder {
	b.network = network
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithAsset(asset string) *PaymentSpendRuleX402ConfigBuilder {
	b.asset = asset
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithAssetTransferMethod(
	method enums.AssetTransferMethod,
) *PaymentSpendRuleX402ConfigBuilder {
	b.assetTransferMethod = &method
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithPolicyId(policyId uuid.UUID) *PaymentSpendRuleX402ConfigBuilder {
	b.policyId = policyId
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) Build() *models.PaymentSpendRuleX402Config {
	return &models.PaymentSpendRuleX402Config{
		PaymentSpendRuleId:  b.paymentSpendRuleId,
		PolicyId:            b.policyId,
		Network:             b.network,
		Asset:               b.asset,
		AssetTransferMethod: b.assetTransferMethod,
	}
}
