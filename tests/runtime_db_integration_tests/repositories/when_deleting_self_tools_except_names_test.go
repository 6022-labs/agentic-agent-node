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

type WhenDeletingSelfToolsExceptNamesTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.SelfToolRepository
}

func WhenDeletingSelfToolsExceptNamesBeforeEach(tc *repositoryTestCase) *WhenDeletingSelfToolsExceptNamesTestingSuite {
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

	return &WhenDeletingSelfToolsExceptNamesTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenDeletingSelfToolsExceptNamesAfterEach(suite *WhenDeletingSelfToolsExceptNamesTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingSelfToolsExceptNames(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given rows outside the kept set", func(t *testing.T) {
				t.Parallel()

				t.Run("Should delete only those, only for that agent", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingSelfToolsExceptNamesBeforeEach(tc)
					defer WhenDeletingSelfToolsExceptNamesAfterEach(suite)

					agentId := uuid.New()
					otherAgentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					seedAgent(suite.dbClient, otherAgentId)

					kept, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("stale").WithEnabled(true).Build())
					assert.NoError(t, err)
					otherAgentRow, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(otherAgentId).WithName("stale").WithEnabled(true).Build())
					assert.NoError(t, err)

					deleted, err := suite.sut.DeleteAllExceptNames(context.Background(), agentId, []string{"bash"})
					assert.NoError(t, err)
					assert.Equal(t, int64(1), deleted)

					got, err := suite.sut.GetById(context.Background(), agentId, kept.Id)
					assert.NoError(t, err)
					assert.NotNil(t, got)

					otherGot, err := suite.sut.GetById(context.Background(), otherAgentId, otherAgentRow.Id)
					assert.NoError(t, err)
					assert.NotNil(t, otherGot)

					var count int64
					suite.dbClient.GetDB().Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})

			t.Run("Given every row is in the kept set", func(t *testing.T) {
				t.Parallel()

				t.Run("Should delete nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingSelfToolsExceptNamesBeforeEach(tc)
					defer WhenDeletingSelfToolsExceptNamesAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("curl").WithEnabled(false).Build())
					assert.NoError(t, err)

					deleted, err := suite.sut.DeleteAllExceptNames(context.Background(), agentId, []string{"bash", "curl"})
					assert.NoError(t, err)
					assert.Equal(t, int64(0), deleted)

					var count int64
					suite.dbClient.GetDB().Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})

			t.Run("Given no names to keep", func(t *testing.T) {
				t.Parallel()

				t.Run("Should delete every row of the agent", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingSelfToolsExceptNamesBeforeEach(tc)
					defer WhenDeletingSelfToolsExceptNamesAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("curl").WithEnabled(false).Build())
					assert.NoError(t, err)

					deleted, err := suite.sut.DeleteAllExceptNames(context.Background(), agentId, nil)
					assert.NoError(t, err)
					assert.Equal(t, int64(2), deleted)

					var count int64
					suite.dbClient.GetDB().Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(0), count)
				})
			})
		})
	}
}
