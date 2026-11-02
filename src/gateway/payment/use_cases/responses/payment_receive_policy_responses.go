package responses

import (
	"encoding/json"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
)

// GetPaymentReceivePolicyResponse is the policy head returned by /config/payment-receive-policy.
type GetPaymentReceivePolicyResponse struct {
	Id                 string    `json:"id"`
	AgentId            string    `json:"agentId"`
	AgentWalletAddress string    `json:"agentWalletAddress"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

func ToGetPaymentReceivePolicyResponse(p *models.ExternalPaymentReceivePolicyPayload) *GetPaymentReceivePolicyResponse {
	if p == nil {
		return nil
	}
	return &GetPaymentReceivePolicyResponse{
		Id:                 p.Id,
		AgentId:            p.AgentId,
		AgentWalletAddress: p.AgentWalletAddress,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}
}

// PaymentReceiveRuleResponse is the published rule shape; the external model doubles as a request body.
type PaymentReceiveRuleResponse struct {
	Id     string                                   `json:"id"`
	Name   string                                   `json:"name"`
	Method string                                   `json:"method"`
	Prices []models.ExternalPaymentReceiveRulePrice `json:"prices"`
	Config json.RawMessage                          `json:"config" swaggertype:"object"`
}

func ToPaymentReceiveRuleResponse(rule *models.ExternalPaymentReceiveRule) *PaymentReceiveRuleResponse {
	if rule == nil {
		return nil
	}
	return &PaymentReceiveRuleResponse{
		Id:     rule.Id,
		Name:   rule.Name,
		Method: rule.Method,
		Prices: rule.Prices,
		Config: rule.Config,
	}
}

// ListPaymentReceiveRulesResponse is the wire shape of /config/payment-receive-policy/rules.
// Rules carry no facilitator URL, so the model is returned directly.
type ListPaymentReceiveRulesResponse struct {
	Rules []PaymentReceiveRuleResponse `json:"rules"`
}

func ToListPaymentReceiveRulesResponse(
	rules []models.ExternalPaymentReceiveRule,
) *ListPaymentReceiveRulesResponse {
	out := make([]PaymentReceiveRuleResponse, 0, len(rules))
	for i := range rules {
		out = append(out, *ToPaymentReceiveRuleResponse(&rules[i]))
	}
	return &ListPaymentReceiveRulesResponse{Rules: out}
}
