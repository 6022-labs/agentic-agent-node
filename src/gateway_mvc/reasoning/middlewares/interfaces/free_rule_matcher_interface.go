package interfaces

import (
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/ethereum/go-ethereum/common"
)

// FreeRuleMatcherInterface reports whether an address matches a zero-price rule.
type FreeRuleMatcherInterface interface {
	MatchesFreeRule(rules []models.ExternalPaymentRule, address common.Address) bool
}
