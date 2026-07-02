package enums

import (
	"fmt"
	"strings"
)

// AddressType is a chain family the agent holds a wallet for; the value doubles as the per-type Vault path segment (wallet/<type>).
type AddressType string

const (
	AddressTypeEvm AddressType = "evm"
	AddressTypeBtc AddressType = "btc"
)

func GetAllAddressTypes() []AddressType {
	return []AddressType{AddressTypeEvm, AddressTypeBtc}
}

func (t AddressType) String() string {
	return string(t)
}

func (t *AddressType) Set(value string) error {
	normalized := AddressType(strings.ToLower(strings.TrimSpace(value)))
	for _, v := range GetAllAddressTypes() {
		if normalized == v {
			*t = v
			return nil
		}
	}
	return fmt.Errorf("invalid AddressType: %s", value)
}
