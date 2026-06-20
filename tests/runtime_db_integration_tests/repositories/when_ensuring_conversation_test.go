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

type WhenEnsuringConversationTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenEnsuringConversationBeforeEach(tc *repositoryTestCase) *WhenEnsuringConversationTestingSuite {
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

	return &WhenEnsuringConversationTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenEnsuringConversationAfterEach(suite *WhenEnsuringConversationTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenEnsuringConversation(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the same conversation is ensured twice", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep a single row via ON CONFLICT DO NOTHING", func(t *testing.T) {
					t.Parallel()
					suite := WhenEnsuringConversationBeforeEach(tc)
					defer WhenEnsuringConversationAfterEach(suite)

					agentId := uuid.New()
					conversation := model_builders.NewConversationBuilder().WithId("conv-dup").WithAgentId(agentId).Build()

					created, err := suite.sut.Ensure(context.Background(), conversation)
					assert.NoError(t, err)
					assert.True(t, created)

					created, err = suite.sut.Ensure(context.Background(), conversation)
					assert.NoError(t, err)
					assert.False(t, created)

					list, err := suite.sut.ListByAgentId(context.Background(), agentId)
					assert.NoError(t, err)
					assert.Len(t, list, 1)
				})
			})
		})
	}
}
