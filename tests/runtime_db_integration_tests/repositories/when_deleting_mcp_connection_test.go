package repositories_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenDeletingMCPConnectionTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.MCPConnectionRepository
}

func WhenDeletingMCPConnectionBeforeEach(tc *repositoryTestCase) *WhenDeletingMCPConnectionTestingSuite {
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

	return &WhenDeletingMCPConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenDeletingMCPConnectionAfterEach(suite *WhenDeletingMCPConnectionTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingMCPConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a non-existent Id", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingMCPConnectionBeforeEach(tc)
					defer WhenDeletingMCPConnectionAfterEach(suite)

					assert.NoError(t, suite.sut.Delete(uuid.New()))
				})
			})

			t.Run("Given an existing connection", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove the row", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingMCPConnectionBeforeEach(tc)
					defer WhenDeletingMCPConnectionAfterEach(suite)

					inserted, err := suite.sut.Create(newMCPConnection("alpha", "https://a.example"))
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.Delete(inserted.Id))

					got, err := suite.sut.GetById(inserted.Id)
					assert.NoError(t, err)
					assert.Nil(t, got)

					var count int64
					suite.dbClient.GetDB().Model(&models.MCPConnection{}).Count(&count)
					assert.Equal(t, int64(0), count)
				})
			})
		})
	}
}
