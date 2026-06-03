package enums

import (
	"fmt"
	"strings"
)

// TransactionStatus is the on-chain lifecycle shared by the mint and ENS flows.
type TransactionStatus string
type TransactionStatusList []TransactionStatus

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusSubmitted TransactionStatus = "submitted"
	TransactionStatusConfirmed TransactionStatus = "confirmed"
	TransactionStatusFailed    TransactionStatus = "failed"
)

func GetAllTransactionStatuses() TransactionStatusList {
	return TransactionStatusList{
		TransactionStatusPending,
		TransactionStatusSubmitted,
		TransactionStatusConfirmed,
		TransactionStatusFailed,
	}
}

func (s TransactionStatus) String() string {
	return string(s)
}

func (l TransactionStatusList) String() string {
	out := make([]string, 0, len(l))
	for _, v := range l {
		out = append(out, string(v))
	}
	return strings.Join(out, ", ")
}

func (s *TransactionStatus) Set(value string) error {
	normalizedValue := strings.ToLower(value)
	for _, v := range GetAllTransactionStatuses() {
		if normalizedValue == v.String() {
			*s = v
			return nil
		}
	}
	return fmt.Errorf("invalid TransactionStatus: %s", value)
}

func (l TransactionStatusList) Contains(s TransactionStatus) bool {
	for _, v := range l {
		if v == s {
			return true
		}
	}
	return false
}
