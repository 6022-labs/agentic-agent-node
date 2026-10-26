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

type WhenUpsertingAgentGuidanceTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentGuidanceRepository
}

func WhenUpsertingAgentGuidanceBeforeEach(tc *common_tests.DatabaseTestCase) *WhenUpsertingAgentGuidanceTestingSuite {
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

	return &WhenUpsertingAgentGuidanceTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentGuidanceRepository(dbClient),
	}
}

func WhenUpsertingAgentGuidanceAfterEach(suite *WhenUpsertingAgentGuidanceTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingAgentGuidance(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the new guidance with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentGuidanceBeforeEach(tc)
					defer WhenUpsertingAgentGuidanceAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					inserted, err := suite.sut.Upsert(context.Background(), &models.AgentGuidance{
						AgentId: agentId,
						Content: "answer concisely",
					})

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)
				})

				t.Run("Should return the new guidance with its content", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentGuidanceBeforeEach(tc)
					defer WhenUpsertingAgentGuidanceAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					inserted, err := suite.sut.Upsert(context.Background(), &models.AgentGuidance{
						AgentId: agentId,
						Content: "answer concisely",
					})

					assert.NoError(t, err)
					assert.Equal(t, "answer concisely", inserted.Content)
				})

				t.Run("Should store a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentGuidanceBeforeEach(tc)
					defer WhenUpsertingAgentGuidanceAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), &models.AgentGuidance{
						AgentId: agentId,
						Content: "answer concisely",
					})
					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.AgentGuidance{}).Count(&count)

					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given a row already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should preserve its Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentGuidanceBeforeEach(tc)
					defer WhenUpsertingAgentGuidanceAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					first, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentGuidanceBuilder().
						WithAgentId(agentId).
						WithContent("first guidance").
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentGuidanceBuilder().
						WithAgentId(agentId).
						WithContent("second guidance").
						Build())

					assert.NoError(t, err)
					assert.Equal(t, first.Id, second.Id)
				})

				t.Run("Should update the content", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentGuidanceBeforeEach(tc)
					defer WhenUpsertingAgentGuidanceAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentGuidanceBuilder().
						WithAgentId(agentId).
						WithContent("first guidance").
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentGuidanceBuilder().
						WithAgentId(agentId).
						WithContent("second guidance").
						Build())

					assert.NoError(t, err)
					assert.Equal(t, "second guidance", second.Content)
				})

				t.Run("Should keep a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentGuidanceBeforeEach(tc)
					defer WhenUpsertingAgentGuidanceAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentGuidanceBuilder().
						WithAgentId(agentId).
						WithContent("first guidance").
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Upsert(context.Background(), model_builders.NewAgentGuidanceBuilder().
						WithAgentId(agentId).
						WithContent("second guidance").
						Build())

					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.AgentGuidance{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
