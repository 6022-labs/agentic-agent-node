package models

import "time"

// ExternalPaymentRule mirrors a runtime rule: atomic-unit Price ("0" = free),
// CAIP-2 Network, AddressPatterns regexes (".*" = everyone). Id is empty in
// create/update bodies. The facilitator is network-level config, fetched
// separately (ExternalPaymentFacilitatorsPayload), not carried on the rule.
type ExternalPaymentRule struct {
	Id                string   `json:"id,omitempty"`
	Name              string   `json:"name"`
	Network           string   `json:"network"`
	Asset             string   `json:"asset"`
	PayTo             string   `json:"payTo"`
	Price             string   `json:"price"`
	MaxTimeoutSeconds int      `json:"maxTimeoutSeconds"`
	Eip712Name        string   `json:"eip712Name"`
	Eip712Version     string   `json:"eip712Version"`
	AddressPatterns   []string `json:"addressPatterns"`
}

// ExternalPaymentFacilitator mirrors one network's runtime-configured x402
// facilitator.
type ExternalPaymentFacilitator struct {
	Network string `json:"network"`
	BaseUrl string `json:"baseUrl"`
}

// ExternalPaymentFacilitatorsPayload mirrors GET /internal/config/payment-facilitators.
type ExternalPaymentFacilitatorsPayload struct {
	Facilitators []ExternalPaymentFacilitator `json:"facilitators"`
}

// ExternalPaymentPolicyPayload mirrors the runtime's policy head; rules live
// on their own endpoint.
type ExternalPaymentPolicyPayload struct {
	Id                 string    `json:"id"`
	AgentId            string    `json:"agentId"`
	AgentWalletAddress string    `json:"agentWalletAddress"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ExternalListPaymentRulesPayload mirrors GET /internal/config/payment-policy/rules.
type ExternalListPaymentRulesPayload struct {
	Rules []ExternalPaymentRule `json:"rules"`
}
