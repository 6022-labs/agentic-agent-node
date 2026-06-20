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

type WhenListingConversationsByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenListingConversationsByAgentIdBeforeEach(tc *repositoryTestCase) *WhenListingConversationsByAgentIdTestingSuite {
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

	return &WhenListingConversationsByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenListingConversationsByAgentIdAfterEach(suite *WhenListingConversationsByAgentIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingConversationsByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given two agents each own conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should list only the queried agent's conversations", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

					agentA := uuid.New()
					agentB := uuid.New()
					_, ensureErr := suite.sut.Ensure(context.Background(),
						model_builders.NewConversationBuilder().WithId("a-1").WithAgentId(agentA).Build())
					assert.NoError(t, ensureErr)
					_, ensureErr = suite.sut.Ensure(context.Background(),
						model_builders.NewConversationBuilder().WithId("a-2").WithAgentId(agentA).Build())
					assert.NoError(t, ensureErr)
					_, ensureErr = suite.sut.Ensure(context.Background(),
						model_builders.NewConversationBuilder().WithId("b-1").WithAgentId(agentB).Build())
					assert.NoError(t, ensureErr)

					list, err := suite.sut.ListByAgentId(context.Background(), agentA)
					assert.NoError(t, err)
					assert.Len(t, list, 2)
					for _, conversation := range list {
						assert.Equal(t, agentA, conversation.AgentId)
					}
				})
			})
		})
	}
}
