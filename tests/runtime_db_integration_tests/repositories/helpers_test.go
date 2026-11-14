package repositories_test

import (
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// newMigratedDatabase creates a fresh migrated database and returns its builder and name for teardown plus its client.
func newMigratedDatabase(
	tc *common_tests.DatabaseTestCase,
) (container_builders.DatabaseContainerBuilderInterface, string, *runtime_db.DatabaseClient) {
	databaseContainerBuilder := tc.ClientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := runtime_db.NewDatabaseClient(
		&settings.DatabaseSettings{Driver: tc.Driver, Url: connectionString},
		&settings.MigrationsSettings{FolderPath: *migrationFolderPath},
		zap.NewNop(),
	)
	if err != nil {
		panic(err)
	}
	if err := dbClient.Migrate(); err != nil {
		panic(err)
	}

	return databaseContainerBuilder, dbName, dbClient
}

// assertUniqueViolation matches the driver-specific duplicate-key wording of postgres, mysql/mariadb and sqlite.
func assertUniqueViolation(t *testing.T, err error) {
	t.Helper()
	if assert.Error(t, err) {
		msg := strings.ToLower(err.Error())
		assert.True(
			t,
			strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique"),
			"not a unique violation: %v",
			err,
		)
	}
}
