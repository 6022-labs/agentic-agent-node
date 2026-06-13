package responses

// ListPaymentRulesResponse is the wire shape of
// /internal/config/payment-policy/rules.
type ListPaymentRulesResponse struct {
	Rules []PaymentRuleResponse `json:"rules"`
}
