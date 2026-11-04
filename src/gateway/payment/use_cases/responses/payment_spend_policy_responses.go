package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// GetPaymentSpendPolicyResponse is the policy head returned by /config/payment-spend-policy.
type GetPaymentSpendPolicyResponse struct {
	Id            string    `json:"id"`
	AgentId       string    `json:"agentId"`
	PayToPatterns []string  `json:"payToPatterns"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func ToGetPaymentSpendPolicyResponse(p *models.ExternalPaymentSpendPolicyPayload) *GetPaymentSpendPolicyResponse {
	if p == nil {
		return nil
	}
	return &GetPaymentSpendPolicyResponse{
		Id:            p.Id,
		AgentId:       p.AgentId,
		PayToPatterns: p.PayToPatterns,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

// PaymentSpendRuleResponse is the published rule shape; the external model doubles as a request body.
type PaymentSpendRuleResponse struct {
	Id                  string                           `json:"id"`
	Name                string                           `json:"name"`
	MaxAmountPerPayment *string                          `json:"maxAmountPerPayment,omitempty"`
	Windows             []PaymentSpendRuleWindowResponse `json:"windows"`
	Method              string                           `json:"method"`
	Config              json.RawMessage                  `json:"config"                        swaggertype:"object"`
}

// PaymentSpendRuleWindowResponse is one window's budget and its spend this period.
type PaymentSpendRuleWindowResponse struct {
	WindowSeconds int        `json:"windowSeconds"`
	MaxAmount     string     `json:"maxAmount"`
	Spent         string     `json:"spent"`
	ResetsAt      *time.Time `json:"resetsAt,omitempty"`
}

func ToPaymentSpendRuleResponse(rule *models.ExternalPaymentSpendRule) *PaymentSpendRuleResponse {
	if rule == nil {
		return nil
	}
	windows := make([]PaymentSpendRuleWindowResponse, 0, len(rule.Windows))
	for _, window := range rule.Windows {
		windows = append(windows, PaymentSpendRuleWindowResponse{
			WindowSeconds: window.WindowSeconds,
			MaxAmount:     window.MaxAmount,
			Spent:         window.Spent,
			ResetsAt:      window.ResetsAt,
		})
	}
	return &PaymentSpendRuleResponse{
		Id:                  rule.Id,
		Name:                rule.Name,
		MaxAmountPerPayment: rule.MaxAmountPerPayment,
		Windows:             windows,
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
