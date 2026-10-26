package responses

// ListPaymentReceiveRulesResponse is the wire shape of
// /internal/config/payment-receive-policy/rules.
type ListPaymentReceiveRulesResponse struct {
	Rules []PaymentReceiveRuleResponse `json:"rules"`
}
