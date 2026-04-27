package repositories_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenGettingLLMConnectionTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.LLMConnectionRepository
}

func WhenGettingLLMConnectionBeforeEach(tc *repositoryTestCase) *WhenGettingLLMConnectionTestingSuite {
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

	return &WhenGettingLLMConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewLLMConnectionRepository(dbClient),
	}
}

func WhenGettingLLMConnectionAfterEach(suite *WhenGettingLLMConnectionTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func ptr(s string) *string { return &s }

func TestWhenGettingLLMConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no connection has been configured", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingLLMConnectionBeforeEach(tc)
					defer WhenGettingLLMConnectionAfterEach(suite)

					got, err := suite.sut.Get()
					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})

			t.Run("Given a connection has been configured", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the persisted row", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingLLMConnectionBeforeEach(tc)
					defer WhenGettingLLMConnectionAfterEach(suite)

					inserted, err := suite.sut.Upsert(&models.LLMConnection{
						Platform:    reasoning_enums.LLMConnectionPlatformOpenAI,
						Model:       "gpt-4o",
						BaseUrl:     ptr("https://api.openai.com/v1"),
						SetupStatus: reasoning_enums.LLMConnectionSetupStatusEstablished,
					})
					assert.NoError(t, err)

					got, err := suite.sut.Get()
					assert.NoError(t, err)
					assert.Equal(t, inserted.Id, got.Id)
					assert.Equal(t, reasoning_enums.LLMConnectionPlatformOpenAI, got.Platform)
					assert.Equal(t, "gpt-4o", got.Model)
					assert.Equal(t, "https://api.openai.com/v1", *got.BaseUrl)
				})
			})
		})
	}
}
