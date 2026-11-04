package models

import (
	"encoding/json"
	"time"
)

// ExternalPaymentSpendRule mirrors a runtime spend rule; Id is empty in create/update bodies.
type ExternalPaymentSpendRule struct {
	Id                  string                           `json:"id,omitempty"`
	Name                string                           `json:"name"`
	MaxAmountPerPayment *string                          `json:"maxAmountPerPayment,omitempty"`
	Windows             []ExternalPaymentSpendRuleWindow `json:"windows"`
	Method              string                           `json:"method"`
	Config              json.RawMessage                  `json:"config"                        swaggertype:"object"`
}

// ExternalPaymentSpendRuleWindow mirrors one window of a runtime spend rule; Spent and ResetsAt are read-only.
type ExternalPaymentSpendRuleWindow struct {
	WindowSeconds int        `json:"windowSeconds"`
	MaxAmount     string     `json:"maxAmount"`
	Spent         string     `json:"spent,omitempty"`
	ResetsAt      *time.Time `json:"resetsAt,omitempty"`
}

// ExternalPaymentSpendPolicyPayload mirrors the runtime's spend policy head.
type ExternalPaymentSpendPolicyPayload struct {
	Id            string    `json:"id"`
	AgentId       string    `json:"agentId"`
	PayToPatterns []string  `json:"payToPatterns"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ExternalUpdatePaymentSpendPolicyPayload mirrors PUT /internal/config/payment-spend-policy.
type ExternalUpdatePaymentSpendPolicyPayload struct {
	PayToPatterns []string `json:"payToPatterns"`
}

// ExternalListPaymentSpendRulesPayload mirrors GET /internal/config/payment-spend-policy/rules.
type ExternalListPaymentSpendRulesPayload struct {
	Rules []ExternalPaymentSpendRule `json:"rules"`
}
