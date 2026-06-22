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

type WhenDeletingLLMConnectionByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.LLMConnectionRepository
}

func WhenDeletingLLMConnectionByAgentIdBeforeEach(tc *repositoryTestCase) *WhenDeletingLLMConnectionByAgentIdTestingSuite {
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

	return &WhenDeletingLLMConnectionByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewLLMConnectionRepository(dbClient),
	}
}

func WhenDeletingLLMConnectionByAgentIdAfterEach(suite *WhenDeletingLLMConnectionByAgentIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingLLMConnectionByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingLLMConnectionByAgentIdBeforeEach(tc)
					defer WhenDeletingLLMConnectionByAgentIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), uuid.New()))
				})
			})

			t.Run("Given a row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should clear the row", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingLLMConnectionByAgentIdBeforeEach(tc)
					defer WhenDeletingLLMConnectionByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentId).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAI).
						WithModel("gpt-4o").
						WithSetupStatus(reasoning_enums.LLMConnectionSetupStatusEstablished).
						Build())

					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), agentId))

					var count int64
					suite.dbClient.GetDB().Model(&models.LLMConnection{}).Count(&count)
					assert.Equal(t, int64(0), count)
				})
			})
		})
	}
}
