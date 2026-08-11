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

type WhenCreatingSelfToolTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.SelfToolRepository
}

func WhenCreatingSelfToolBeforeEach(tc *repositoryTestCase) *WhenCreatingSelfToolTestingSuite {
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

	return &WhenCreatingSelfToolTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenCreatingSelfToolAfterEach(suite *WhenCreatingSelfToolTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingSelfTool(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a valid tool", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist it with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingSelfToolBeforeEach(tc)
					defer WhenCreatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithId(uuid.Nil).WithAgentId(agentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)
					assert.Equal(t, "bash", inserted.Name)
					assert.True(t, inserted.Enabled)

					var count int64
					suite.dbClient.GetDB().Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given the agent already has a tool with that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should reject the duplicate and keep the original row", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingSelfToolBeforeEach(tc)
					defer WhenCreatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithId(uuid.Nil).WithAgentId(agentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithId(uuid.Nil).WithAgentId(agentId).WithName("bash").WithEnabled(false).Build())
					assert.Error(t, err)

					var count int64
					suite.dbClient.GetDB().Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given the same name under a different agent", func(t *testing.T) {
				t.Parallel()

				t.Run("Should allow it since uniqueness is scoped per agent", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingSelfToolBeforeEach(tc)
					defer WhenCreatingSelfToolAfterEach(suite)

					firstAgentId := uuid.New()
					secondAgentId := uuid.New()
					seedAgent(suite.dbClient, firstAgentId)
					seedAgent(suite.dbClient, secondAgentId)

					_, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithId(uuid.Nil).WithAgentId(firstAgentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithId(uuid.Nil).WithAgentId(secondAgentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB().Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})
		})
	}
}
