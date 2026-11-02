package models

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// PaymentReceiveRuleX402Config is the token a rule prices; its transfer method and EIP-712 domain are read on-chain.
type PaymentReceiveRuleX402Config struct {
	PaymentReceiveRuleId uuid.UUID          `gorm:"primaryKey;size:36"                                                        json:"-"`
	PaymentReceiveRule   PaymentReceiveRule `gorm:"foreignKey:PaymentReceiveRuleId;references:Id;constraint:OnDelete:CASCADE" json:"-"`
	// Copied from the rule so the database can hold one rule per token.
	PolicyId uuid.UUID `gorm:"not null;size:36;uniqueIndex:uq_payment_receive_rule_x402_configs_policy_asset,priority:1" json:"-"`

	Network             string                    `gorm:"not null;size:64;uniqueIndex:uq_payment_receive_rule_x402_configs_policy_asset,priority:2" json:"network"`
	Asset               string                    `gorm:"not null;size:42;uniqueIndex:uq_payment_receive_rule_x402_configs_policy_asset,priority:3" json:"asset"`
	PayTo               string                    `gorm:"not null;size:42"                                                                          json:"payTo"`
	AssetTransferMethod enums.AssetTransferMethod `gorm:"not null;size:16"                                                                          json:"assetTransferMethod"`
	Eip712Name          *string                   `gorm:"size:64"                                                                                   json:"eip712Name,omitempty"`
	Eip712Version       *string                   `gorm:"size:16"                                                                                   json:"eip712Version,omitempty"`
}

func (PaymentReceiveRuleX402Config) TableName() string {
	return "payment_receive_rule_x402_configs"
}

// Parse decodes an x402 config blob into the receiver, erroring on
// malformed/empty input. A method (not a package func) so each method's config
// type owns its own Parse without colliding in the models package.
func (c *PaymentReceiveRuleX402Config) Parse(raw []byte) error {
	if len(raw) == 0 {
		return errors.NewValidationError("config", "is required for x402 rules")
	}
	if err := json.Unmarshal(raw, c); err != nil {
		return errors.NewValidationError("config", "must be a valid x402 config object")
	}
	return nil
}

// Marshal encodes the config as the wire blob (persistence fields are omitted).
func (c *PaymentReceiveRuleX402Config) Marshal() ([]byte, error) {
	return json.Marshal(c)
}
