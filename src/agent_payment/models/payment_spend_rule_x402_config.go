package models

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// PaymentSpendRuleX402Config matches a 402 quote; it stores patterns, not concrete values.
type PaymentSpendRuleX402Config struct {
	PaymentSpendRuleId uuid.UUID        `gorm:"primaryKey;size:36"                                                      json:"-"`
	PaymentSpendRule   PaymentSpendRule `gorm:"foreignKey:PaymentSpendRuleId;references:Id;constraint:OnDelete:CASCADE" json:"-"`

	NetworkPatterns []string `gorm:"serializer:json;not null" json:"networkPatterns"`
	AssetPatterns   []string `gorm:"serializer:json;not null" json:"assetPatterns"`
	PayToPatterns   []string `gorm:"serializer:json;not null" json:"payToPatterns"`
	// permit2 grants a standing max-uint256 allowance, so a rule must be able to demand one-shot eip3009.
	AssetTransferMethodPatterns []string `gorm:"serializer:json;not null" json:"assetTransferMethodPatterns"`
}

func (PaymentSpendRuleX402Config) TableName() string {
	return "payment_spend_rule_x402_configs"
}

// Matches reports whether every facet of the quote is in scope; all four arrays must hit.
func (c *PaymentSpendRuleX402Config) Matches(network, asset, payTo, assetTransferMethod string) bool {
	return c.matchesPatterns(c.NetworkPatterns, network) &&
		c.matchesPatterns(c.AssetPatterns, asset) &&
		c.matchesPatterns(c.PayToPatterns, payTo) &&
		c.matchesPatterns(c.AssetTransferMethodPatterns, assetTransferMethod)
}

// matchesPatterns full-matches case-insensitively; lowercasing the pattern would rewrite \D into \d.
func (c *PaymentSpendRuleX402Config) matchesPatterns(patterns []string, value string) bool {
	value = strings.TrimSpace(value)
	for _, pattern := range patterns {
		compiled, err := regexp.Compile("^(?i:" + strings.TrimSpace(pattern) + ")$")
		if err != nil {
			continue
		}
		if compiled.MatchString(value) {
			return true
		}
	}
	return false
}

// Parse decodes an x402 spend config blob into the receiver.
func (c *PaymentSpendRuleX402Config) Parse(raw []byte) error {
	if len(raw) == 0 {
		return errors.NewValidationError("config", "is required for x402 spend rules")
	}
	if err := json.Unmarshal(raw, c); err != nil {
		return errors.NewValidationError("config", "must be a valid x402 spend config object")
	}
	return nil
}

// Marshal encodes the config as the wire blob (persistence fields are omitted).
func (c *PaymentSpendRuleX402Config) Marshal() ([]byte, error) {
	return json.Marshal(c)
}
