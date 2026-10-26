package common_tests

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
)

// DatabaseTestCase is the per-driver matrix entry shared by every suite that needs a database.
type DatabaseTestCase struct {
	Name            string
	Driver          enums.DatabaseDriver
	ClientBuilderFn func() container_builders.DatabaseContainerBuilderInterface
}

func DatabaseTestCases() []*DatabaseTestCase {
	return []*DatabaseTestCase{
		{
			Name:   "PostgreSQL",
			Driver: enums.DatabaseDriverPostgres,
			ClientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
				return container_builders.GetPostgresContainerBuilder()
			},
		},
		{
			Name:   "MySQL",
			Driver: enums.DatabaseDriverMySQL,
			ClientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
				return container_builders.GetMySQLContainerBuilder()
			},
		},
		{
			Name:   "MariaDB",
			Driver: enums.DatabaseDriverMariaDB,
			ClientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
				return container_builders.GetMariaDBContainerBuilder()
			},
		},
		{
			Name:   "SQLite",
			Driver: enums.DatabaseDriverSqlite,
			ClientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
				return container_builders.GetSqliteContainerBuilder()
			},
		},
	}
}
