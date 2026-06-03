package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenCreatingMCPConnectionTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.MCPConnectionRepository
}

func WhenCreatingMCPConnectionBeforeEach(tc *repositoryTestCase) *WhenCreatingMCPConnectionTestingSuite {
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

	return &WhenCreatingMCPConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenCreatingMCPConnectionAfterEach(suite *WhenCreatingMCPConnectionTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func newMCPConnection(name, url string, headerKeys ...string) *models.MCPConnection {
	return model_builders.NewMCPConnectionBuilder().
		WithName(name).
		WithUrl(url).
		WithHeaderAuthKeys(headerKeys).
		Build()
}

func TestWhenCreatingMCPConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a valid connection", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist it with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingMCPConnectionBeforeEach(tc)
					defer WhenCreatingMCPConnectionAfterEach(suite)

					inserted, err := suite.sut.Create(context.Background(), newMCPConnection("github", "https://mcp.github.example", "Authorization"))
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)
					assert.Equal(t, "github", inserted.Name)

					var count int64
					suite.dbClient.GetDB().Model(&models.MCPConnection{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given a connection with no header auth keys", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist an empty JSON array", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingMCPConnectionBeforeEach(tc)
					defer WhenCreatingMCPConnectionAfterEach(suite)

					inserted, err := suite.sut.Create(context.Background(), newMCPConnection("anon", "https://mcp.anon.example"))
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), inserted.Id)
					assert.NoError(t, err)
					assert.NotNil(t, got)
					assert.Empty(t, got.HeaderAuthKeys)
				})
			})
		})
	}
}
