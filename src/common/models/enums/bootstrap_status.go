package enums

import "strings"

type BootstrapStatus string
type BootstrapStatusList []BootstrapStatus

const (
	BootstrapStatusAwaitingVault BootstrapStatus = "awaiting-vault"
	BootstrapStatusReady         BootstrapStatus = "ready"
)

func GetAllBootstrapStatuses() BootstrapStatusList {
	return BootstrapStatusList{
		BootstrapStatusAwaitingVault,
		BootstrapStatusReady,
	}
}

func (k BootstrapStatusList) String() string {
	var values []string
	for _, v := range k {
		values = append(values, string(v))
	}
	return strings.Join(values, ", ")
}

func (list BootstrapStatusList) Contains(value BootstrapStatus) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
