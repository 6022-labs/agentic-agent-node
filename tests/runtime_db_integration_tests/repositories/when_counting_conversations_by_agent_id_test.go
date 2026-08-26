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

type WhenCountingConversationsByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenCountingConversationsByAgentIdBeforeEach(
	tc *repositoryTestCase,
) *WhenCountingConversationsByAgentIdTestingSuite {
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

	return &WhenCountingConversationsByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenCountingConversationsByAgentIdAfterEach(suite *WhenCountingConversationsByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCountingConversationsByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the agent has conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should count only its own", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingConversationsByAgentIdBeforeEach(tc)
					defer WhenCountingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					otherAgentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					seedAgent(suite.dbClient, otherAgentId)

					for _, id := range []string{"a", "b", "c"} {
						_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
							WithId(id).
							WithAgentId(agentId).
							Build())
						assert.NoError(t, err)
					}
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("other").
						WithAgentId(otherAgentId).
						Build())
					assert.NoError(t, err)

					total, err := suite.sut.CountByAgentId(ctx, agentId)

					assert.NoError(t, err)
					assert.Equal(t, uint(3), total)
				})
			})

			t.Run("Given the agent has no conversation", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return zero", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingConversationsByAgentIdBeforeEach(tc)
					defer WhenCountingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					total, err := suite.sut.CountByAgentId(ctx, agentId)

					assert.NoError(t, err)
					assert.Equal(t, uint(0), total)
				})
			})
		})
	}
}
