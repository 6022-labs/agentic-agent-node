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

func WhenCreatingSelfToolBeforeEach(tc *common_tests.DatabaseTestCase) *WhenCreatingSelfToolTestingSuite {
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

	return &WhenCreatingSelfToolTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenCreatingSelfToolAfterEach(suite *WhenCreatingSelfToolTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingSelfTool(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a valid tool", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist it with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingSelfToolBeforeEach(tc)
					defer WhenCreatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					inserted, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithId(uuid.Nil).
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)
					assert.Equal(t, "bash", inserted.Name)
					assert.True(t, inserted.Enabled)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given the agent already has a tool with that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should reject the duplicate", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingSelfToolBeforeEach(tc)
					defer WhenCreatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithId(uuid.Nil).
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithId(uuid.Nil).
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(false).
							Build(),
					)

					assertUniqueViolation(t, err)
				})

				t.Run("Should keep the original row", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingSelfToolBeforeEach(tc)
					defer WhenCreatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithId(uuid.Nil).
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					_, _ = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithId(uuid.Nil).
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(false).
							Build(),
					)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.SelfTool{}).Count(&count)
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
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(firstAgentId).Build()).Error; err != nil {
						panic(err)
					}
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(secondAgentId).Build()).Error; err != nil {
						panic(err)
					}

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithId(uuid.Nil).
							WithAgentId(firstAgentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithId(uuid.Nil).
							WithAgentId(secondAgentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})
		})
	}
}
