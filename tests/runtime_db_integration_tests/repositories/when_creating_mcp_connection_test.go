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
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
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

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("github").
							WithUrl("https://mcp.github.example").
							WithHeaderAuthKeys([]string{"Authorization"}).
							Build(),
					)
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

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("anon").
							WithUrl("https://mcp.anon.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), inserted.AgentId, inserted.Id)
					assert.NoError(t, err)
					assert.NotNil(t, got)
					assert.Empty(t, got.HeaderAuthKeys)
				})
			})

			t.Run("Given the agent already has a connection with that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should reject the duplicate", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingMCPConnectionBeforeEach(tc)
					defer WhenCreatingMCPConnectionAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("github").
							WithUrl("https://mcp.github.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)
					assert.NoError(t, err)

					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("github").
							WithUrl("https://other.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)

					assertUniqueViolation(t, err)
				})

				t.Run("Should keep the original row", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingMCPConnectionBeforeEach(tc)
					defer WhenCreatingMCPConnectionAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("github").
							WithUrl("https://mcp.github.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)
					assert.NoError(t, err)

					_, _ = suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("github").
							WithUrl("https://other.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)

					var count int64
					suite.dbClient.GetDB().Model(&models.MCPConnection{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given the same name under a different agent", func(t *testing.T) {
				t.Parallel()

				t.Run("Should allow it since uniqueness is scoped per agent", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingMCPConnectionBeforeEach(tc)
					defer WhenCreatingMCPConnectionAfterEach(suite)

					firstAgentId := uuid.New()
					secondAgentId := uuid.New()
					seedAgent(suite.dbClient, firstAgentId)
					seedAgent(suite.dbClient, secondAgentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(firstAgentId).
							WithName("github").
							WithUrl("https://a.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)
					assert.NoError(t, err)

					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(secondAgentId).
							WithName("github").
							WithUrl("https://b.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)
					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB().Model(&models.MCPConnection{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})
		})
	}
}
