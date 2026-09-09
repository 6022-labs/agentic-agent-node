package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// GetPaymentSpendPolicyResponse is the policy head returned by /config/payment-spend-policy.
type GetPaymentSpendPolicyResponse struct {
	Id        string    `json:"id"`
	AgentId   string    `json:"agentId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ToGetPaymentSpendPolicyResponse(p *models.ExternalPaymentSpendPolicyPayload) *GetPaymentSpendPolicyResponse {
	if p == nil {
		return nil
	}
	return &GetPaymentSpendPolicyResponse{
		Id:        p.Id,
		AgentId:   p.AgentId,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// PaymentSpendRuleResponse is the published rule shape; the external model doubles as a request body.
type PaymentSpendRuleResponse struct {
	Id                  string          `json:"id"`
	Name                string          `json:"name"`
	MaxAmountPerPayment string          `json:"maxAmountPerPayment"`
	WindowSeconds       int             `json:"windowSeconds"`
	MaxAmountPerWindow  string          `json:"maxAmountPerWindow"`
	WindowSpent         string          `json:"windowSpent"`
	WindowResetsAt      *time.Time      `json:"windowResetsAt,omitempty"`
	Method              string          `json:"method"`
	Config              json.RawMessage `json:"config"                   swaggertype:"object"`
}

func ToPaymentSpendRuleResponse(rule *models.ExternalPaymentSpendRule) *PaymentSpendRuleResponse {
	if rule == nil {
		return nil
	}
	return &PaymentSpendRuleResponse{
		Id:                  rule.Id,
		Name:                rule.Name,
		MaxAmountPerPayment: rule.MaxAmountPerPayment,
		WindowSeconds:       rule.WindowSeconds,
		MaxAmountPerWindow:  rule.MaxAmountPerWindow,
		WindowSpent:         rule.WindowSpent,
		WindowResetsAt:      rule.WindowResetsAt,
		Method:              rule.Method,
		Config:              rule.Config,
	}
}

// ListPaymentSpendRulesResponse is the wire shape of /config/payment-spend-policy/rules.
type ListPaymentSpendRulesResponse struct {
	Rules []PaymentSpendRuleResponse `json:"rules"`
}

func ToListPaymentSpendRulesResponse(rules []models.ExternalPaymentSpendRule) *ListPaymentSpendRulesResponse {
	out := make([]PaymentSpendRuleResponse, 0, len(rules))
	for i := range rules {
		out = append(out, *ToPaymentSpendRuleResponse(&rules[i]))
	}
	return &ListPaymentSpendRulesResponse{Rules: out}
}
