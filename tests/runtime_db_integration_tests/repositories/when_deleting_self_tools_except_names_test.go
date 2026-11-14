package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenDeletingSelfToolsExceptNamesTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.SelfToolRepository
}

func WhenDeletingSelfToolsExceptNamesBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingSelfToolsExceptNamesTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenDeletingSelfToolsExceptNamesTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenDeletingSelfToolsExceptNamesAfterEach(suite *WhenDeletingSelfToolsExceptNamesTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingSelfToolsExceptNames(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given rows outside the kept set", func(t *testing.T) {
				t.Parallel()

				t.Run("Should delete only the rows outside the kept set", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingSelfToolsExceptNamesBeforeEach(tc)
					defer WhenDeletingSelfToolsExceptNamesAfterEach(suite)

					agentId := uuid.New()
					otherAgentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					seeds.SeedAgent(suite.dbClient, otherAgentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("stale").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(otherAgentId).
							WithName("stale").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					deleted, err := suite.sut.DeleteAllExceptNames(context.Background(), agentId, []string{"bash"})
					assert.NoError(t, err)
					assert.Equal(t, int64(1), deleted)
				})

				t.Run("Should keep the rows in the kept set", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingSelfToolsExceptNamesBeforeEach(tc)
					defer WhenDeletingSelfToolsExceptNamesAfterEach(suite)

					agentId := uuid.New()
					otherAgentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					seeds.SeedAgent(suite.dbClient, otherAgentId)

					kept, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("stale").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(otherAgentId).
							WithName("stale").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					_, err = suite.sut.DeleteAllExceptNames(context.Background(), agentId, []string{"bash"})
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), agentId, kept.Id)

					assert.NoError(t, err)
					assert.NotNil(t, got)
				})

				t.Run("Should leave other agents rows untouched", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingSelfToolsExceptNamesBeforeEach(tc)
					defer WhenDeletingSelfToolsExceptNamesAfterEach(suite)

					agentId := uuid.New()
					otherAgentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					seeds.SeedAgent(suite.dbClient, otherAgentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("stale").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					otherAgentRow, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(otherAgentId).
							WithName("stale").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					_, err = suite.sut.DeleteAllExceptNames(context.Background(), agentId, []string{"bash"})
					assert.NoError(t, err)

					otherGot, err := suite.sut.GetById(context.Background(), otherAgentId, otherAgentRow.Id)

					assert.NoError(t, err)
					assert.NotNil(t, otherGot)
				})
			})

			t.Run("Given every row is in the kept set", func(t *testing.T) {
				t.Parallel()

				t.Run("Should delete nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingSelfToolsExceptNamesBeforeEach(tc)
					defer WhenDeletingSelfToolsExceptNamesAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("curl").
							WithEnabled(false).
							Build(),
					)
					assert.NoError(t, err)

					deleted, err := suite.sut.DeleteAllExceptNames(
						context.Background(),
						agentId,
						[]string{"bash", "curl"},
					)
					assert.NoError(t, err)
					assert.Equal(t, int64(0), deleted)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.SelfTool{}).Count(&count)
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
					seeds.SeedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("curl").
							WithEnabled(false).
							Build(),
					)
					assert.NoError(t, err)

					deleted, err := suite.sut.DeleteAllExceptNames(context.Background(), agentId, nil)
					assert.NoError(t, err)
					assert.Equal(t, int64(2), deleted)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.SelfTool{}).Count(&count)
					assert.Equal(t, int64(0), count)
				})
			})
		})
	}
}
