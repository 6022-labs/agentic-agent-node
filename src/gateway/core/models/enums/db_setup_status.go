package enums

import "strings"

type DbSetupStatus string
type DbSetupStatusList []DbSetupStatus

const (
	DbSetupStatusConfigured   DbSetupStatus = "configured"
	DbSetupStatusEstablishing DbSetupStatus = "establishing"
	DbSetupStatusEstablished  DbSetupStatus = "established"
	DbSetupStatusFailed       DbSetupStatus = "failed"
)

func GetAllDbSetupStatuses() DbSetupStatusList {
	return DbSetupStatusList{
		DbSetupStatusConfigured,
		DbSetupStatusEstablishing,
		DbSetupStatusEstablished,
		DbSetupStatusFailed,
	}
}

func (k DbSetupStatusList) String() string {
	var values []string
	for _, v := range k {
		values = append(values, string(v))
	}
	return strings.Join(values, ", ")
}

func (list DbSetupStatusList) Contains(value DbSetupStatus) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
