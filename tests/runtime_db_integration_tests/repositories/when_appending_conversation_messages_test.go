package repositories_test

import (
	"context"
	"testing"
	"time"

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

type WhenAppendingConversationMessagesTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationMessageRepository
	conversationRepository   *repositories.ConversationRepository
}

func WhenAppendingConversationMessagesBeforeEach(
	tc *repositoryTestCase,
) *WhenAppendingConversationMessagesTestingSuite {
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

	return &WhenAppendingConversationMessagesTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationMessageRepository(dbClient),
		conversationRepository:   repositories.NewConversationRepository(dbClient),
	}
}

func WhenAppendingConversationMessagesAfterEach(suite *WhenAppendingConversationMessagesTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func newMessage(conversationId, content string) models.ConversationMessage {
	return model_builders.NewConversationMessageBuilder().
		WithConversationId(conversationId).
		WithRole("user").
		WithContent(content).
		WithTimestamp(time.Now().UTC().Truncate(time.Microsecond)).
		Build()
}

// ensureConversations seeds the conversation heads that messages reference,
// satisfying the conversation_messages.conversation_id FK.
func ensureConversations(
	dbClient *runtime_db.DatabaseClient,
	conversationRepository *repositories.ConversationRepository,
	conversationIds ...string,
) {
	for _, id := range conversationIds {
		agentId := uuid.New()
		seedAgent(dbClient, agentId)
		if _, err := conversationRepository.Ensure(context.Background(), model_builders.NewConversationBuilder().WithId(id).WithAgentId(agentId).Build()); err != nil {
			panic(err)
		}
	}
}

func TestWhenAppendingConversationMessages(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given an empty input", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenAppendingConversationMessagesBeforeEach(tc)
					defer WhenAppendingConversationMessagesAfterEach(suite)

					inserted, err := suite.sut.AppendMessages(context.Background(), nil)
					assert.NoError(t, err)
					assert.Empty(t, inserted)
				})
			})

			t.Run("Given two distinct new messages", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert both", func(t *testing.T) {
					t.Parallel()
					suite := WhenAppendingConversationMessagesBeforeEach(tc)
					defer WhenAppendingConversationMessagesAfterEach(suite)

					ensureConversations(suite.dbClient, suite.conversationRepository, "conv-1")
					inserted, err := suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{
						newMessage("conv-1", "hello"),
						newMessage("conv-1", "world"),
					})

					assert.NoError(t, err)
					assert.Len(t, inserted, 2)

					var count int64
					suite.dbClient.GetDB().Model(&models.ConversationMessage{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})

			t.Run("Given a hash that already exists is re-inserted (race-safety net)", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep DB count correct via ON CONFLICT DO NOTHING", func(t *testing.T) {
					t.Parallel()
					suite := WhenAppendingConversationMessagesBeforeEach(tc)
					defer WhenAppendingConversationMessagesAfterEach(suite)

					ensureConversations(suite.dbClient, suite.conversationRepository, "conv-2")
					_, err := suite.sut.AppendMessages(
						context.Background(),
						[]models.ConversationMessage{newMessage("conv-2", "hi")},
					)
					assert.NoError(t, err)

					_, err = suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{
						newMessage("conv-2", "hi"),
						newMessage("conv-2", "different content"),
					})

					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB().Model(&models.ConversationMessage{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})

			t.Run("Given a batch with two semantically identical messages", func(t *testing.T) {
				t.Parallel()

				t.Run("Should land only one row thanks to ON CONFLICT DO NOTHING", func(t *testing.T) {
					t.Parallel()
					suite := WhenAppendingConversationMessagesBeforeEach(tc)
					defer WhenAppendingConversationMessagesAfterEach(suite)

					ensureConversations(suite.dbClient, suite.conversationRepository, "conv-3")
					_, err := suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{
						newMessage("conv-3", "same"),
						newMessage("conv-3", "same"),
					})

					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB().Model(&models.ConversationMessage{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

		})
	}
}
