package repositories_test

import (
	"context"
	"testing"

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

type WhenGettingMCPConnectionByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.MCPConnectionRepository
}

func WhenGettingMCPConnectionByIdBeforeEach(tc *repositoryTestCase) *WhenGettingMCPConnectionByIdTestingSuite {
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

	return &WhenGettingMCPConnectionByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenGettingMCPConnectionByIdAfterEach(suite *WhenGettingMCPConnectionByIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingMCPConnectionById(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a non-existent Id", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingMCPConnectionByIdBeforeEach(tc)
					defer WhenGettingMCPConnectionByIdAfterEach(suite)

					got, err := suite.sut.GetById(context.Background(), uuid.New(), uuid.New())
					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})

			t.Run("Given a connection has been created", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the persisted row by Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingMCPConnectionByIdBeforeEach(tc)
					defer WhenGettingMCPConnectionByIdAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(context.Background(), model_builders.NewMCPConnectionBuilder().WithAgentId(agentId).WithName("github").WithUrl("https://mcp.github.example").WithHeaderAuthKeys([]string{"Authorization"}).Build())
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), inserted.AgentId, inserted.Id)
					assert.NoError(t, err)
					assert.Equal(t, inserted.Id, got.Id)
					assert.Equal(t, "github", got.Name)
					assert.Equal(t, "https://mcp.github.example", got.Url)
					assert.Len(t, got.HeaderAuthKeys, 1)
					assert.Equal(t, "Authorization", string(got.HeaderAuthKeys[0]))
				})
			})
		})
	}
}
