package repositories_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenDeletingLLMConnectionTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.LLMConnectionRepository
}

func WhenDeletingLLMConnectionBeforeEach(tc *repositoryTestCase) *WhenDeletingLLMConnectionTestingSuite {
	databaseContainerBuilder := tc.clientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := runtime_db.NewDatabaseClient(
		&settings.DatabaseSettings{Driver: tc.driver, Url: connectionString},
		&settings.MigrationsSettings{FolderPath: *migrationFolderPath},
		zap.NewNop(),
	)
	if err != nil {
		panic(err)
	}
	if err := dbClient.Migrate(); err != nil {
		panic(err)
	}

	return &WhenDeletingLLMConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewLLMConnectionRepository(dbClient),
	}
}

func WhenDeletingLLMConnectionAfterEach(suite *WhenDeletingLLMConnectionTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingLLMConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingLLMConnectionBeforeEach(tc)
					defer WhenDeletingLLMConnectionAfterEach(suite)

					assert.NoError(t, suite.sut.Delete())
				})
			})

			t.Run("Given a row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should clear the row", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingLLMConnectionBeforeEach(tc)
					defer WhenDeletingLLMConnectionAfterEach(suite)

					_, err := suite.sut.Upsert(&models.LLMConnection{
						Platform:    reasoning_enums.LLMConnectionPlatformOpenAI,
						Model:       "gpt-4o",
						SetupStatus: reasoning_enums.LLMConnectionSetupStatusEstablished,
					})
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.Delete())

					var count int64
					suite.dbClient.GetDB().Model(&models.LLMConnection{}).Count(&count)
					assert.Equal(t, int64(0), count)
				})
			})
		})
	}
}
