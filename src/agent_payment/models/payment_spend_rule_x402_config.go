package models

import (
	"encoding/json"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// PaymentSpendRuleX402Config names one token concretely, since an ERC-20 address is only unique per network.
type PaymentSpendRuleX402Config struct {
	PaymentSpendRuleId uuid.UUID        `gorm:"primaryKey;size:36"                                                      json:"-"`
	PaymentSpendRule   PaymentSpendRule `gorm:"foreignKey:PaymentSpendRuleId;references:Id;constraint:OnDelete:CASCADE" json:"-"`
	// Copied from the rule so the database can hold one rule per token.
	PolicyId uuid.UUID `gorm:"not null;size:36;uniqueIndex:uq_payment_spend_rule_x402_configs_policy_asset,priority:1" json:"-"`

	Network string `gorm:"not null;size:64;uniqueIndex:uq_payment_spend_rule_x402_configs_policy_asset,priority:2" json:"network"`
	Asset   string `gorm:"not null;size:42;uniqueIndex:uq_payment_spend_rule_x402_configs_policy_asset,priority:3" json:"asset"`
	// nil accepts any, EIP-2612 Permit included; eip3009 refuses permit2's standing max-uint256 allowance.
	AssetTransferMethod *enums.AssetTransferMethod `gorm:"size:16" json:"assetTransferMethod,omitempty"`
}

func (PaymentSpendRuleX402Config) TableName() string {
	return "payment_spend_rule_x402_configs"
}

func (c *PaymentSpendRuleX402Config) AcceptsAssetTransferMethod(assetTransferMethod string) bool {
	return c.AssetTransferMethod == nil ||
		strings.EqualFold(strings.TrimSpace(assetTransferMethod), c.AssetTransferMethod.String())
}

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
