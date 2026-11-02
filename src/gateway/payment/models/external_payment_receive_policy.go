package models

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/common/x402"
)

// ExternalPaymentReceiveRule mirrors a runtime rule, one per token; Id is empty in create/update bodies.
type ExternalPaymentReceiveRule struct {
	Id     string                            `json:"id,omitempty"`
	Name   string                            `json:"name"`
	Method string                            `json:"method"`
	Prices []ExternalPaymentReceiveRulePrice `json:"prices"`
	Config json.RawMessage                   `json:"config"       swaggertype:"object"`
}

// ExternalPaymentReceiveRulePrice mirrors one price of a runtime rule and the payers it applies to.
type ExternalPaymentReceiveRulePrice struct {
	Price           string   `json:"price"`
	AddressPatterns []string `json:"addressPatterns"`
}

// ExternalPaymentReceivePolicyPayload mirrors the runtime's policy head.
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

// x402 access outcomes, mirroring the runtime's enum.
const (
	PaymentReceiveAccessOutcomeFree            = "free"
	PaymentReceiveAccessOutcomePayable         = "payable"
	PaymentReceiveAccessOutcomePaymentRequired = "payment_required"
)

// ExternalPaymentReceiveAccessDecision mirrors POST /internal/payment/receive/access.
type ExternalPaymentReceiveAccessDecision struct {
	Outcome    string                            `json:"outcome"`
	Payer      string                            `json:"payer,omitempty"`
	ErrorCode  string                            `json:"errorCode,omitempty"`
	Accepts    []x402.PaymentRequirements        `json:"accepts"`
	Settlement *ExternalPaymentReceiveSettlement `json:"settlement,omitempty"`
}

// ExternalPaymentReceiveSettlement is what verifying and settling a payable request needs.
type ExternalPaymentReceiveSettlement struct {
	Requirements       x402.PaymentRequirements `json:"requirements"`
	FacilitatorBaseUrl string                   `json:"facilitatorBaseUrl"`
	Network            string                   `json:"network"`
}
