package models

import (
	"encoding/json"
	"time"
)

// ExternalPaymentReceiveRule mirrors a runtime rule: method-neutral Name/Price/Method
// plus a method-specific Config blob. Id is empty in create/update bodies. The
// facilitator is network-level config fetched separately
// (ExternalPaymentFacilitatorsPayload), not carried on the rule.
type ExternalPaymentReceiveRule struct {
	Id     string          `json:"id,omitempty"`
	Name   string          `json:"name"`
	Price  string          `json:"price"`
	Method string          `json:"method"`
	Config json.RawMessage `json:"config"       swaggertype:"object"`
}

// ExternalPaymentFacilitator mirrors one network's runtime-configured x402
// facilitator.
type ExternalPaymentFacilitator struct {
	Network           string `json:"network"`
	BaseUrl           string `json:"baseUrl"`
	MaxTimeoutSeconds int    `json:"maxTimeoutSeconds"`
}

// ExternalPaymentFacilitatorsPayload mirrors GET /internal/config/payment-facilitators.
type ExternalPaymentFacilitatorsPayload struct {
	Facilitators []ExternalPaymentFacilitator `json:"facilitators"`
}

// ExternalPaymentReceivePolicyPayload mirrors the runtime's policy head; rules live
// on their own endpoint.
type ExternalPaymentReceivePolicyPayload struct {
	Id                 string    `json:"id"`
	AgentId            string    `json:"agentId"`
	AgentWalletAddress string    `json:"agentWalletAddress"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ExternalListPaymentReceiveRulesPayload mirrors GET /internal/config/payment-receive-policy/rules.
type ExternalListPaymentReceiveRulesPayload struct {
	Rules []ExternalPaymentReceiveRule `json:"rules"`
}
