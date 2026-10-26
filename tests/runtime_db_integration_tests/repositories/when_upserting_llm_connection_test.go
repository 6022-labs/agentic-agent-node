package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
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

type WhenUpsertingLLMConnectionTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.LLMConnectionRepository
}

func WhenUpsertingLLMConnectionBeforeEach(tc *common_tests.DatabaseTestCase) *WhenUpsertingLLMConnectionTestingSuite {
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

	return &WhenUpsertingLLMConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewLLMConnectionRepository(dbClient),
	}
}

func WhenUpsertingLLMConnectionAfterEach(suite *WhenUpsertingLLMConnectionTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingLLMConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the new connection with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAi).
						WithModel("gpt-4o").
						WithBaseUrl("https://api.openai.com/v1").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)
				})

				t.Run("Should store a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAi).
						WithModel("gpt-4o").
						WithBaseUrl("https://api.openai.com/v1").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())
					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.LLMConnection{}).Count(&count)

					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given a row already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should preserve its Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					first, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAi).
						WithModel("gpt-4o").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformAnthropic).
						WithModel("claude-opus-4-7").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)
					assert.Equal(t, first.Id, second.Id)
				})

				t.Run("Should replace its platform", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAi).
						WithModel("gpt-4o").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformAnthropic).
						WithModel("claude-opus-4-7").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformAnthropic, second.Platform)
				})

				t.Run("Should replace its model", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAi).
						WithModel("gpt-4o").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformAnthropic).
						WithModel("claude-opus-4-7").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)
					assert.Equal(t, "claude-opus-4-7", second.Model)
				})

				t.Run("Should keep a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAi).
						WithModel("gpt-4o").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformAnthropic).
						WithModel("claude-opus-4-7").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.LLMConnection{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
