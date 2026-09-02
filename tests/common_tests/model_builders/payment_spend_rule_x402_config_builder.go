package model_builders

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/google/uuid"
)

type PaymentSpendRuleX402ConfigBuilder struct {
	paymentSpendRuleId          uuid.UUID
	networkPatterns             []string
	assetPatterns               []string
	payToPatterns               []string
	assetTransferMethodPatterns []string
}

// NewPaymentSpendRuleX402ConfigBuilder defaults to the seeded catch-all: every facet open.
func NewPaymentSpendRuleX402ConfigBuilder() *PaymentSpendRuleX402ConfigBuilder {
	return &PaymentSpendRuleX402ConfigBuilder{
		paymentSpendRuleId:          uuid.New(),
		networkPatterns:             []string{".*"},
		assetPatterns:               []string{".*"},
		payToPatterns:               []string{".*"},
		assetTransferMethodPatterns: []string{".*"},
	}
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithPaymentSpendRuleId(
	id uuid.UUID,
) *PaymentSpendRuleX402ConfigBuilder {
	b.paymentSpendRuleId = id
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithNetworkPatterns(
	patterns []string,
) *PaymentSpendRuleX402ConfigBuilder {
	b.networkPatterns = patterns
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithAssetPatterns(
	patterns []string,
) *PaymentSpendRuleX402ConfigBuilder {
	b.assetPatterns = patterns
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithPayToPatterns(
	patterns []string,
) *PaymentSpendRuleX402ConfigBuilder {
	b.payToPatterns = patterns
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) WithAssetTransferMethodPatterns(
	patterns []string,
) *PaymentSpendRuleX402ConfigBuilder {
	b.assetTransferMethodPatterns = patterns
	return b
}

func (b *PaymentSpendRuleX402ConfigBuilder) Build() *models.PaymentSpendRuleX402Config {
	return &models.PaymentSpendRuleX402Config{
		PaymentSpendRuleId:          b.paymentSpendRuleId,
		NetworkPatterns:             b.networkPatterns,
		AssetPatterns:               b.assetPatterns,
		PayToPatterns:               b.payToPatterns,
		AssetTransferMethodPatterns: b.assetTransferMethodPatterns,
	}
}
