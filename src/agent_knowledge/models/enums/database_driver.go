package enums

import "strings"

type DatabaseDriver string
type DatabaseDriverList []DatabaseDriver

const (
	DatabaseDriverPostgres DatabaseDriver = "postgres"
	DatabaseDriverMySQL    DatabaseDriver = "mysql"
	DatabaseDriverMariaDB  DatabaseDriver = "mariadb"
)

func GetAllDatabaseDrivers() DatabaseDriverList {
	return DatabaseDriverList{
		DatabaseDriverPostgres,
		DatabaseDriverMySQL,
		DatabaseDriverMariaDB,
	}
}

func (d DatabaseDriver) String() string {
	return string(d)
}

// ToGooseDialect maps to the dialect names goose understands (mariadb is
// handled by the mysql goose dialect).
func (d DatabaseDriver) ToGooseDialect() string {
	switch d {
	case DatabaseDriverPostgres:
		return "postgres"
	case DatabaseDriverMySQL, DatabaseDriverMariaDB:
		return "mysql"
	default:
		return ""
	}
}

func (l DatabaseDriverList) String() string {
	out := make([]string, 0, len(l))
	for _, d := range l {
		out = append(out, string(d))
	}
	return strings.Join(out, ", ")
}

func (l DatabaseDriverList) Contains(d DatabaseDriver) bool {
	for _, v := range l {
		if v == d {
			return true
		}
	}
	return false
}
