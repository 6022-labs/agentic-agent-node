package repositories_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenListingMCPConnectionsTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.MCPConnectionRepository
}

func WhenListingMCPConnectionsBeforeEach(tc *repositoryTestCase) *WhenListingMCPConnectionsTestingSuite {
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

	return &WhenListingMCPConnectionsTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenListingMCPConnectionsAfterEach(suite *WhenListingMCPConnectionsTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingMCPConnections(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no connections exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty slice", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingMCPConnectionsBeforeEach(tc)
					defer WhenListingMCPConnectionsAfterEach(suite)

					list, err := suite.sut.List()
					assert.NoError(t, err)
					assert.Empty(t, list)
				})
			})

			t.Run("Given multiple connections were created", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them ordered by created_at ASC", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingMCPConnectionsBeforeEach(tc)
					defer WhenListingMCPConnectionsAfterEach(suite)

					_, err := suite.sut.Create(newMCPConnection("alpha", "https://a.example"))
					assert.NoError(t, err)
					_, err = suite.sut.Create(newMCPConnection("beta", "https://b.example"))
					assert.NoError(t, err)

					list, err := suite.sut.List()
					assert.NoError(t, err)
					assert.Len(t, list, 2)
					assert.Equal(t, "alpha", list[0].Name)
					assert.Equal(t, "beta", list[1].Name)
				})
			})
		})
	}
}
