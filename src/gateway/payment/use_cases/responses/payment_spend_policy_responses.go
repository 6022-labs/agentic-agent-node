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

// PaymentSpendRuleResponse is one spend rule as the gateway publishes it; the
// external model doubles as a request body, so it is not the response shape.
type PaymentSpendRuleResponse struct {
	Id                  string          `json:"id"`
	Name                string          `json:"name"`
	MaxAmountPerPayment string          `json:"maxAmountPerPayment"`
	Method              string          `json:"method"`
	Config              json.RawMessage `json:"config"              swaggertype:"object"`
}

func ToPaymentSpendRuleResponse(rule *models.ExternalPaymentSpendRule) *PaymentSpendRuleResponse {
	if rule == nil {
		return nil
	}
	return &PaymentSpendRuleResponse{
		Id:                  rule.Id,
		Name:                rule.Name,
		MaxAmountPerPayment: rule.MaxAmountPerPayment,
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
