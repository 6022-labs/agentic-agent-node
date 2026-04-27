package migrations_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenMigratingMigrationsTestCase struct {
	name            string
	driver          enums.DatabaseDriver
	clientBuilderFn func() container_builders.DatabaseContainerBuilderInterface
}

var WhenMigratingMigrationsTestCases = []*WhenMigratingMigrationsTestCase{
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
}

type WhenMigratingMigrationsTestingSuite struct {
	dbName                   string
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *runtime_db.DatabaseClient
}

func WhenMigratingMigrationsBeforeEach(testCase *WhenMigratingMigrationsTestCase) *WhenMigratingMigrationsTestingSuite {
	databaseContainerBuilder := testCase.clientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := runtime_db.NewDatabaseClient(
		&settings.DatabaseSettings{Driver: testCase.driver, Url: connectionString},
		&settings.MigrationsSettings{FolderPath: *migrationFolderPath},
		zap.NewNop(),
	)
	if err != nil {
		panic(err)
	}

	return &WhenMigratingMigrationsTestingSuite{
		dbName:                   dbName,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      dbClient,
	}
}

func WhenMigratingMigrationsAfterEach(suite *WhenMigratingMigrationsTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenMigratingMigrations(t *testing.T) {
	t.Parallel()

	for _, testCase := range WhenMigratingMigrationsTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no migrations are applied", func(t *testing.T) {
				t.Parallel()

				t.Run("Should apply all migrations without error", func(t *testing.T) {
					t.Parallel()

					suite := WhenMigratingMigrationsBeforeEach(testCase)
					defer WhenMigratingMigrationsAfterEach(suite)

					assert.NoError(t, suite.sut.Migrate())
				})
			})

			t.Run("Given migrations are already applied", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be idempotent on a second run", func(t *testing.T) {
					t.Parallel()

					suite := WhenMigratingMigrationsBeforeEach(testCase)
					defer WhenMigratingMigrationsAfterEach(suite)

					assert.NoError(t, suite.sut.Migrate())
					assert.NoError(t, suite.sut.Migrate())
				})
			})
		})
	}
}
