package enums

import "strings"

type PaymentReceiveAccessOutcome string
type PaymentReceiveAccessOutcomeList []PaymentReceiveAccessOutcome

const (
	// A matching rule prices this caller at zero.
	PaymentReceiveAccessOutcomeFree PaymentReceiveAccessOutcome = "free"
	// The signature covers a matching paid rule and is ready to settle.
	PaymentReceiveAccessOutcomePayable PaymentReceiveAccessOutcome = "payable"
	// Refused, carrying what the caller may pay instead.
	PaymentReceiveAccessOutcomePaymentRequired PaymentReceiveAccessOutcome = "payment_required"
)

func GetAllPaymentReceiveAccessOutcomes() PaymentReceiveAccessOutcomeList {
	return PaymentReceiveAccessOutcomeList{
		PaymentReceiveAccessOutcomeFree,
		PaymentReceiveAccessOutcomePayable,
		PaymentReceiveAccessOutcomePaymentRequired,
	}
}

func (o PaymentReceiveAccessOutcome) String() string {
	return string(o)
}

func (o PaymentReceiveAccessOutcomeList) String() string {
	out := make([]string, 0, len(o))
	for _, v := range o {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}
