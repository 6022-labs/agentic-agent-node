package model_builders

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/google/uuid"
)

type PaymentSpendRuleX402ConfigBuilder struct {
	paymentSpendRuleId  uuid.UUID
	network             string
	asset               string
	assetTransferMethod *enums.AssetTransferMethod
	payToPatterns       []string
}

// NewPaymentSpendRuleX402ConfigBuilder names one token on one chain, payable to anyone.
func NewPaymentSpendRuleX402ConfigBuilder() *PaymentSpendRuleX402ConfigBuilder {
	return &PaymentSpendRuleX402ConfigBuilder{
		paymentSpendRuleId: uuid.New(),
		network:            "eip155:80002",
		asset:              "0x41e94eb019c0762f9bfcf9fb1e58725bfb0e7582",
		payToPatterns:      []string{".*"},
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

func (b *PaymentSpendRuleX402ConfigBuilder) WithPayToPatterns(
	patterns []string,
) *PaymentSpendRuleX402ConfigBuilder {
	b.payToPatterns = patterns
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) Build() *models.PaymentSpendRuleX402Config {
	return &models.PaymentSpendRuleX402Config{
		PaymentSpendRuleId:  b.paymentSpendRuleId,
		Network:             b.network,
		Asset:               b.asset,
		AssetTransferMethod: b.assetTransferMethod,
		PayToPatterns:       b.payToPatterns,
	}
}
