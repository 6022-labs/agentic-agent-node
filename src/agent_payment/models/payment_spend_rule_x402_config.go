package models

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
)

// PaymentSpendRuleX402Config names one token concretely, since an ERC-20 address is only unique per network.
type PaymentSpendRuleX402Config struct {
	PaymentSpendRuleId uuid.UUID        `gorm:"primaryKey;size:36"                                                      json:"-"`
	PaymentSpendRule   PaymentSpendRule `gorm:"foreignKey:PaymentSpendRuleId;references:Id;constraint:OnDelete:CASCADE" json:"-"`

	Network string `gorm:"not null;size:64" json:"network"`
	Asset   string `gorm:"not null;size:42" json:"asset"`
	// nil accepts either; eip3009 refuses permit2's standing max-uint256 allowance.
	AssetTransferMethod *enums.AssetTransferMethod `gorm:"size:16"                  json:"assetTransferMethod,omitempty"`
	PayToPatterns       []string                   `gorm:"serializer:json;not null" json:"payToPatterns"`
}

func (PaymentSpendRuleX402Config) TableName() string {
	return "payment_spend_rule_x402_configs"
}

// Matches reports whether the quote is this rule's money, paid to someone it allows.
func (c *PaymentSpendRuleX402Config) Matches(network, asset, payTo, assetTransferMethod string) bool {
	if !strings.EqualFold(strings.TrimSpace(network), strings.TrimSpace(c.Network)) {
		return false
	}
	if !c.sameAddress(asset, c.Asset) {
		return false
	}
	if c.AssetTransferMethod != nil &&
		!strings.EqualFold(strings.TrimSpace(assetTransferMethod), c.AssetTransferMethod.String()) {
		return false
	}
	return c.matchesPayTo(payTo)
}

// sameAddress compares hex addresses by value, so checksummed and lowercase forms agree.
func (c *PaymentSpendRuleX402Config) sameAddress(left, right string) bool {
	left, right = strings.TrimSpace(left), strings.TrimSpace(right)
	if !common.IsHexAddress(left) || !common.IsHexAddress(right) {
		return false
	}
	return common.HexToAddress(left) == common.HexToAddress(right)
}

// matchesPayTo full-matches case-insensitively; lowercasing the pattern would rewrite \D into \d.
func (c *PaymentSpendRuleX402Config) matchesPayTo(payTo string) bool {
	payTo = strings.TrimSpace(payTo)
	for _, pattern := range c.PayToPatterns {
		compiled, err := regexp.Compile("^(?i:" + strings.TrimSpace(pattern) + ")$")
		if err != nil {
			continue
		}
		if compiled.MatchString(payTo) {
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
