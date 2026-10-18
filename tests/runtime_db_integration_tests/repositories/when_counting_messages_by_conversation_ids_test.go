package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
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

type WhenCountingMessagesByConversationIdsTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationMessageRepository
	conversations            *repositories.ConversationRepository
}

func WhenCountingMessagesByConversationIdsBeforeEach(
	tc *repositoryTestCase,
) *WhenCountingMessagesByConversationIdsTestingSuite {
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

	return &WhenCountingMessagesByConversationIdsTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationMessageRepository(dbClient),
		conversations:            repositories.NewConversationRepository(dbClient),
	}
}

func WhenCountingMessagesByConversationIdsAfterEach(suite *WhenCountingMessagesByConversationIdsTestingSuite) {
	// The pool outlives the dropped database otherwise, and the server runs out of clients.
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCountingMessagesByConversationIds(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given conversations holding messages", func(t *testing.T) {
				t.Parallel()

				t.Run("Should count per conversation", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingMessagesByConversationIdsBeforeEach(tc)
					defer WhenCountingMessagesByConversationIdsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					for _, id := range []string{"busy", "quiet", "empty"} {
						_, err := suite.conversations.Create(ctx, model_builders.NewConversationBuilder().
							WithId(id).
							WithAgentId(agentId).
							Build())
						assert.NoError(t, err)
					}

					_, err := suite.sut.AppendMessages(ctx, []models.ConversationMessage{
						newMessage("busy", "one"),
						newMessage("busy", "two"),
						newMessage("quiet", "only"),
					})
					assert.NoError(t, err)

					counts, err := suite.sut.CountByConversationIds(ctx, []string{"busy", "quiet", "empty"})

					assert.NoError(t, err)
					assert.Equal(t, int64(2), counts["busy"])
					assert.Equal(t, int64(1), counts["quiet"])
				})

				t.Run("Should omit the empty ones", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingMessagesByConversationIdsBeforeEach(tc)
					defer WhenCountingMessagesByConversationIdsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					for _, id := range []string{"busy", "quiet", "empty"} {
						_, err := suite.conversations.Create(ctx, model_builders.NewConversationBuilder().
							WithId(id).
							WithAgentId(agentId).
							Build())
						assert.NoError(t, err)
					}

					_, err := suite.sut.AppendMessages(ctx, []models.ConversationMessage{
						newMessage("busy", "one"),
						newMessage("busy", "two"),
						newMessage("quiet", "only"),
					})
					assert.NoError(t, err)

					counts, err := suite.sut.CountByConversationIds(ctx, []string{"busy", "quiet", "empty"})

					assert.NoError(t, err)
					// A conversation with no message is absent, so the caller reads it as zero.
					_, hasEmpty := counts["empty"]
					assert.False(t, hasEmpty)
				})

				t.Run("Should ignore conversations that were not asked for", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingMessagesByConversationIdsBeforeEach(tc)
					defer WhenCountingMessagesByConversationIdsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					for _, id := range []string{"wanted", "other"} {
						_, err := suite.conversations.Create(ctx, model_builders.NewConversationBuilder().
							WithId(id).
							WithAgentId(agentId).
							Build())
						assert.NoError(t, err)
					}

					_, err := suite.sut.AppendMessages(ctx, []models.ConversationMessage{
						newMessage("wanted", "one"),
						newMessage("other", "two"),
					})
					assert.NoError(t, err)

					counts, err := suite.sut.CountByConversationIds(ctx, []string{"wanted"})

					assert.NoError(t, err)
					assert.Len(t, counts, 1)
					assert.Equal(t, int64(1), counts["wanted"])
				})
			})

			t.Run("Given no id is asked for", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty map without querying", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingMessagesByConversationIdsBeforeEach(tc)
					defer WhenCountingMessagesByConversationIdsAfterEach(suite)

					counts, err := suite.sut.CountByConversationIds(context.Background(), nil)

					assert.NoError(t, err)
					assert.Empty(t, counts)
				})
			})
		})
	}
}
