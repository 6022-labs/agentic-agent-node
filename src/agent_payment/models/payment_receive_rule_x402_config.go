package models

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/google/uuid"
)

// PaymentReceiveRuleX402Config is the x402 method's typed settings for one rule — both
// the DB row (gorm tags) and the wire shape of the config blob (json tags);
// PaymentReceiveRuleId/PaymentReceiveRule are persistence-only (json:"-"). AssetTransferMethod
// + the EIP-712 domain are detected from the asset on-chain at save.
// Eip712Name/Version are nil for permit2, which signs over Permit2's own fixed
// domain rather than the token's.
type PaymentReceiveRuleX402Config struct {
	PaymentReceiveRuleId uuid.UUID                 `gorm:"primaryKey;size:36"                                                        json:"-"`
	PaymentReceiveRule   PaymentReceiveRule        `gorm:"foreignKey:PaymentReceiveRuleId;references:Id;constraint:OnDelete:CASCADE" json:"-"`
	Network              string                    `gorm:"not null;size:64"                                                          json:"network"`
	Asset                string                    `gorm:"not null;size:42"                                                          json:"asset"`
	PayTo                string                    `gorm:"not null;size:42"                                                          json:"payTo"`
	AssetTransferMethod  enums.AssetTransferMethod `gorm:"not null;size:16"                                                          json:"assetTransferMethod"`
	Eip712Name           *string                   `gorm:"size:64"                                                                   json:"eip712Name,omitempty"`
	Eip712Version        *string                   `gorm:"size:16"                                                                   json:"eip712Version,omitempty"`
	AddressPatterns      []string                  `gorm:"serializer:json;not null"                                                  json:"addressPatterns"`
}

// Matches reports whether the rule is open to this address.
func (c *PaymentReceiveRuleX402Config) Matches(address string) bool {
	// Case-sensitive on purpose: folding it would newly match uppercase-hex patterns.
	lowered := strings.ToLower(strings.TrimSpace(address))
	for _, pattern := range c.AddressPatterns {
		compiled, err := regexp.Compile("^(?:" + pattern + ")$")
		if err != nil {
			continue
		}
		if compiled.MatchString(lowered) {
			return true
		}
	}
	return false
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
