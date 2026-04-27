package repositories_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenListingConversationMessagesTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationMessageRepository
}

func WhenListingConversationMessagesBeforeEach(tc *repositoryTestCase) *WhenListingConversationMessagesTestingSuite {
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

	return &WhenListingConversationMessagesTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationMessageRepository(dbClient),
	}
}

func WhenListingConversationMessagesAfterEach(suite *WhenListingConversationMessagesTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingConversationMessages(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no messages were ever appended", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty slice", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationMessagesBeforeEach(tc)
					defer WhenListingConversationMessagesAfterEach(suite)

					messages, err := suite.sut.ListByConversation("missing")
					assert.NoError(t, err)
					assert.Empty(t, messages)
				})
			})

			t.Run("Given messages from two conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return only the requested conversation, ordered by timestamp ASC", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationMessagesBeforeEach(tc)
					defer WhenListingConversationMessagesAfterEach(suite)

					now := time.Now().UTC().Truncate(time.Microsecond)
					_, err := suite.sut.AppendMessages([]models.ConversationMessage{
						{ConversationId: "a", Author: "alice", Role: "user", Content: "a-second", Timestamp: now.Add(time.Second)},
						{ConversationId: "a", Author: "alice", Role: "user", Content: "a-first", Timestamp: now},
						{ConversationId: "b", Author: "bob", Role: "user", Content: "b-only", Timestamp: now},
					})
					assert.NoError(t, err)

					messages, err := suite.sut.ListByConversation("a")
					assert.NoError(t, err)
					assert.Len(t, messages, 2)
					assert.Equal(t, "a-first", messages[0].Content)
					assert.Equal(t, "a-second", messages[1].Content)

					otherMessages, err := suite.sut.ListByConversation("b")
					assert.NoError(t, err)
					assert.Len(t, otherMessages, 1)
					assert.Equal(t, "b-only", otherMessages[0].Content)
				})
			})
		})
	}
}
