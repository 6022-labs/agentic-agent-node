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

func WhenUpsertingLLMConnectionBeforeEach(tc *repositoryTestCase) *WhenUpsertingLLMConnectionTestingSuite {
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

	return &WhenUpsertingLLMConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewLLMConnectionRepository(dbClient),
	}
}

func WhenUpsertingLLMConnectionAfterEach(suite *WhenUpsertingLLMConnectionTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingLLMConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert and return the new connection with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAI).
						WithModel("gpt-4o").
						WithBaseUrl("https://api.openai.com/v1").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)

					var count int64
					suite.dbClient.GetDB().Model(&models.LLMConnection{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given a row already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should update the existing row in place and preserve Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingLLMConnectionBeforeEach(tc)
					defer WhenUpsertingLLMConnectionAfterEach(suite)

					first, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAI).
						WithModel("gpt-4o").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithPlatform(reasoning_enums.LLMConnectionPlatformAnthropic).
						WithModel("claude-opus-4-7").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)
					assert.Equal(t, first.Id, second.Id)
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformAnthropic, second.Platform)
					assert.Equal(t, "claude-opus-4-7", second.Model)

					var count int64
					suite.dbClient.GetDB().Model(&models.LLMConnection{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
