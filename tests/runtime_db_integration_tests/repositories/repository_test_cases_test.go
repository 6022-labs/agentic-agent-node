package repositories_test

import (
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
)

// repositoryTestCase is the per-driver matrix entry shared by every repository suite.
type repositoryTestCase struct {
	name            string
	driver          enums.DatabaseDriver
	clientBuilderFn func() container_builders.DatabaseContainerBuilderInterface
}

var repositoryTestCases = []*repositoryTestCase{
	{
		name:   "PostgreSQL",
		driver: enums.DatabaseDriverPostgres,
		clientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
			return container_builders.GetPostgresContainerBuilder()
		},
	},
	{
		name:   "MySQL",
		driver: enums.DatabaseDriverMySQL,
		clientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
			return container_builders.GetMySQLContainerBuilder()
		},
	},
	{
		name:   "MariaDB",
		driver: enums.DatabaseDriverMariaDB,
		clientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
			return container_builders.GetMariaDBContainerBuilder()
		},
	},
	{
		name:   "SQLite",
		driver: enums.DatabaseDriverSqlite,
		clientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
			return container_builders.GetSqliteContainerBuilder()
		},
	},
}
