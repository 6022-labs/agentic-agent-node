package models

import (
	"encoding/json"

	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// PaymentRuleX402Config is the x402 method's typed settings for one rule — both
// the DB row (gorm tags) and the wire shape of the config blob (json tags);
// PaymentRuleId/PaymentRule are persistence-only (json:"-"). The settlement
// timeout is network-level (facilitator config), not here.
type PaymentRuleX402Config struct {
	PaymentRuleId   uuid.UUID   `gorm:"primaryKey;size:36" json:"-"`
	PaymentRule     PaymentRule `gorm:"foreignKey:PaymentRuleId;references:Id;constraint:OnDelete:CASCADE" json:"-"`
	Network         string      `gorm:"not null;size:64" json:"network"`
	Asset           string      `gorm:"not null;size:42" json:"asset"`
	PayTo           string      `gorm:"not null;size:42" json:"payTo"`
	Eip712Name      string      `gorm:"not null;size:64" json:"eip712Name,omitempty"`
	Eip712Version   string      `gorm:"not null;size:16" json:"eip712Version,omitempty"`
	AddressPatterns []string    `gorm:"serializer:json;not null" json:"addressPatterns"`
}

func (PaymentRuleX402Config) TableName() string {
	return "payment_rule_x402_configs"
}

// Parse decodes an x402 config blob into the receiver, erroring on
// malformed/empty input. A method (not a package func) so each method's config
// type owns its own Parse without colliding in the models package.
func (c *PaymentRuleX402Config) Parse(raw []byte) error {
	if len(raw) == 0 {
		return errors.NewValidationError("config", "is required for x402 rules")
	}
	if err := json.Unmarshal(raw, c); err != nil {
		return errors.NewValidationError("config", "must be a valid x402 config object")
	}
	return nil
}

// Marshal encodes the config as the wire blob (persistence fields are omitted).
func (c *PaymentRuleX402Config) Marshal() ([]byte, error) {
	return json.Marshal(c)
}
