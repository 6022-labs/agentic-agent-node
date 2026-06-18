package enums

import "strings"

type DbDriver string
type DbDriverList []DbDriver

const (
	DbDriverPostgres DbDriver = "postgres"
	DbDriverMySQL    DbDriver = "mysql"
	DbDriverMariaDB  DbDriver = "mariadb"
)

func GetAllDbDrivers() DbDriverList {
	return DbDriverList{
		DbDriverPostgres,
		DbDriverMySQL,
		DbDriverMariaDB,
	}
}

func (k DbDriverList) String() string {
	var values []string
	for _, v := range k {
		values = append(values, string(v))
	}
	return strings.Join(values, ", ")
}

func (list DbDriverList) Contains(value DbDriver) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
