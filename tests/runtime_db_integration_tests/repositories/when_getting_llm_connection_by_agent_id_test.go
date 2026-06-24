package repositories_test

import (
	"context"
	"testing"

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

type WhenGettingLLMConnectionByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.LLMConnectionRepository
}

func WhenGettingLLMConnectionByAgentIdBeforeEach(tc *repositoryTestCase) *WhenGettingLLMConnectionByAgentIdTestingSuite {
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

	return &WhenGettingLLMConnectionByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewLLMConnectionRepository(dbClient),
	}
}

func WhenGettingLLMConnectionByAgentIdAfterEach(suite *WhenGettingLLMConnectionByAgentIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingLLMConnectionByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no connection has been configured", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingLLMConnectionByAgentIdBeforeEach(tc)
					defer WhenGettingLLMConnectionByAgentIdAfterEach(suite)

					got, err := suite.sut.GetByAgentId(context.Background(), uuid.New())
					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})

			t.Run("Given a connection has been configured", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the persisted row", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingLLMConnectionByAgentIdBeforeEach(tc)
					defer WhenGettingLLMConnectionByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAi).
						WithModel("gpt-4o").
						WithBaseUrl("https://api.openai.com/v1").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)

					got, err := suite.sut.GetByAgentId(context.Background(), agentId)
					assert.NoError(t, err)
					assert.Equal(t, inserted.Id, got.Id)
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAi, got.Platform)
					assert.Equal(t, "gpt-4o", got.Model)
					assert.Equal(t, "https://api.openai.com/v1", *got.BaseUrl)
				})
			})
		})
	}
}
