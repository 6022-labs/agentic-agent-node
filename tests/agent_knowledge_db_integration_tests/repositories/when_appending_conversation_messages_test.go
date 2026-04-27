package repositories_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type repositoryTestCase struct {
	name            string
	driver          enums.DatabaseDriver
	clientBuilderFn func() container_builders.DatabaseContainerBuilderInterface
}

var repositoryTestCases = []*repositoryTestCase{
	{
		name:   "PostgreSQL",
		driver: enums.DatabaseDriverPostgres,
		clientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
			return container_builders.GetPostgresContainerBuilder()
		},
	},
	{
		name:   "MySQL",
		driver: enums.DatabaseDriverMySQL,
		clientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
			return container_builders.GetMySQLContainerBuilder()
		},
	},
	{
		name:   "MariaDB",
		driver: enums.DatabaseDriverMariaDB,
		clientBuilderFn: func() container_builders.DatabaseContainerBuilderInterface {
			return container_builders.GetMariaDBContainerBuilder()
		},
	},
}

type WhenAppendingConversationMessagesTestingSuite struct {
	dbName                   string
	dbClient                 *agent_knowledge_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationMessageRepository
}

func WhenAppendingConversationMessagesBeforeEach(tc *repositoryTestCase) *WhenAppendingConversationMessagesTestingSuite {
	databaseContainerBuilder := tc.clientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindAgentKnowledgeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := agent_knowledge_db.NewDatabaseClient(
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
	}
}

func WhenAppendingConversationMessagesAfterEach(suite *WhenAppendingConversationMessagesTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func newMessage(conversationId, content string) models.ConversationMessage {
	return models.ConversationMessage{
		ConversationId: conversationId,
		Author:         "alice.eth",
		Role:           "user",
		Platform:       "slack",
		Content:        content,
		Timestamp:      time.Now().UTC().Truncate(time.Microsecond),
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

					inserted, err := suite.sut.AppendMessages(nil)
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

					inserted, err := suite.sut.AppendMessages([]models.ConversationMessage{
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

			t.Run("Given a message that already exists by hash", func(t *testing.T) {
				t.Parallel()

				t.Run("Should skip the duplicate and only insert the new one", func(t *testing.T) {
					t.Parallel()
					suite := WhenAppendingConversationMessagesBeforeEach(tc)
					defer WhenAppendingConversationMessagesAfterEach(suite)

					_, err := suite.sut.AppendMessages([]models.ConversationMessage{newMessage("conv-2", "hi")})
					assert.NoError(t, err)

					inserted, err := suite.sut.AppendMessages([]models.ConversationMessage{
						newMessage("conv-2", "hi"),
						newMessage("conv-2", "different content"),
					})
					assert.NoError(t, err)
					assert.Len(t, inserted, 1)
					assert.Equal(t, "different content", inserted[0].Content)

					var count int64
					suite.dbClient.GetDB().Model(&models.ConversationMessage{}).Count(&count)
					assert.Equal(t, int64(2), count)
				})
			})

			t.Run("Given a batch with two semantically identical messages", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert only one row", func(t *testing.T) {
					t.Parallel()
					suite := WhenAppendingConversationMessagesBeforeEach(tc)
					defer WhenAppendingConversationMessagesAfterEach(suite)

					_, err := suite.sut.AppendMessages([]models.ConversationMessage{
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
