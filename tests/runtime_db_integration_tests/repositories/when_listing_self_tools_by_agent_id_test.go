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

type WhenListingSelfToolsByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.SelfToolRepository
}

func WhenListingSelfToolsByAgentIdBeforeEach(tc *repositoryTestCase) *WhenListingSelfToolsByAgentIdTestingSuite {
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

	return &WhenListingSelfToolsByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenListingSelfToolsByAgentIdAfterEach(suite *WhenListingSelfToolsByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingSelfToolsByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no tools exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty slice", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingSelfToolsByAgentIdBeforeEach(tc)
					defer WhenListingSelfToolsByAgentIdAfterEach(suite)

					list, err := suite.sut.ListByAgentId(context.Background(), uuid.New())
					assert.NoError(t, err)
					assert.Empty(t, list)
				})
			})

			t.Run("Given tools exist for several agents", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return only that agent's rows ordered by name ASC", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingSelfToolsByAgentIdBeforeEach(tc)
					defer WhenListingSelfToolsByAgentIdAfterEach(suite)

					agentId := uuid.New()
					otherAgentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					seedAgent(suite.dbClient, otherAgentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("web_search").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(false).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(otherAgentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					list, err := suite.sut.ListByAgentId(context.Background(), agentId)
					assert.NoError(t, err)
					assert.Len(t, list, 2)
					assert.Equal(t, "bash", list[0].Name)
					assert.Equal(t, "web_search", list[1].Name)
					assert.Equal(t, agentId, list[0].AgentId)
					assert.Equal(t, agentId, list[1].AgentId)
				})
			})
		})
	}
}
