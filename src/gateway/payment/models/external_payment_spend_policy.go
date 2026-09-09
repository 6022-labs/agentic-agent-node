package models

import (
	"encoding/json"
	"time"
)

// ExternalPaymentSpendRule mirrors a runtime spend rule; Id is empty in create/update bodies.
type ExternalPaymentSpendRule struct {
	Id                  string          `json:"id,omitempty"`
	Name                string          `json:"name"`
	MaxAmountPerPayment string          `json:"maxAmountPerPayment"`
	Method              string          `json:"method"`
	Config              json.RawMessage `json:"config"              swaggertype:"object"`
}

// ExternalPaymentSpendPolicyPayload mirrors the runtime's spend policy head.
type ExternalPaymentSpendPolicyPayload struct {
	Id        string    `json:"id"`
	AgentId   string    `json:"agentId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ExternalListPaymentSpendRulesPayload mirrors GET /internal/config/payment-spend-policy/rules.
type ExternalListPaymentSpendRulesPayload struct {
	Rules []ExternalPaymentSpendRule `json:"rules"`
}
