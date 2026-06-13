package enums

import (
	"fmt"
	"strings"
)

type PaymentMethod string
type PaymentMethodList []PaymentMethod

const (
	PaymentMethodX402 PaymentMethod = "x402"
)

func GetAllPaymentMethods() PaymentMethodList {
	return PaymentMethodList{
		PaymentMethodX402,
	}
}

func (m PaymentMethod) String() string {
	return string(m)
}

func (m PaymentMethodList) String() string {
	out := make([]string, 0, len(m))
	for _, v := range m {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}

func (m *PaymentMethod) Set(value string) error {
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	for _, v := range GetAllPaymentMethods() {
		if normalizedValue == v.String() {
			*m = v
			return nil
		}
	}
	return fmt.Errorf("invalid PaymentMethod: %s", value)
}

func (m PaymentMethodList) Contains(method PaymentMethod) bool {
	for _, v := range m {
		if v == method {
			return true
		}
	}
	return false
}
