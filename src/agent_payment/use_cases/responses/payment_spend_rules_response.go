package responses

// ListPaymentSpendRulesResponse is the wire shape of /internal/config/payment-spend-policy/rules.
type ListPaymentSpendRulesResponse struct {
	Rules []PaymentSpendRuleResponse `json:"rules"`
}
