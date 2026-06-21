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
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenListingConversationMessagesByConversationIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationMessageRepository
	conversationRepository   *repositories.ConversationRepository
}

func WhenListingConversationMessagesByConversationIdBeforeEach(tc *repositoryTestCase) *WhenListingConversationMessagesByConversationIdTestingSuite {
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

	return &WhenListingConversationMessagesByConversationIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationMessageRepository(dbClient),
		conversationRepository:   repositories.NewConversationRepository(dbClient),
	}
}

func WhenListingConversationMessagesByConversationIdAfterEach(suite *WhenListingConversationMessagesByConversationIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingConversationMessagesByConversationId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no messages were ever appended", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty slice", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationMessagesByConversationIdBeforeEach(tc)
					defer WhenListingConversationMessagesByConversationIdAfterEach(suite)

					messages, err := suite.sut.ListByConversationId(context.Background(), "missing")
					assert.NoError(t, err)
					assert.Empty(t, messages)
				})
			})

			t.Run("Given messages from two conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return only the requested conversation, ordered by timestamp ASC", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationMessagesByConversationIdBeforeEach(tc)
					defer WhenListingConversationMessagesByConversationIdAfterEach(suite)

					ensureConversations(suite.conversationRepository, "a", "b")
					now := time.Now().UTC().Truncate(time.Microsecond)
					_, err := suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{
						model_builders.NewConversationMessageBuilder().WithConversationId("a").WithContent("a-second").WithTimestamp(now.Add(time.Second)).Build(),
						model_builders.NewConversationMessageBuilder().WithConversationId("a").WithContent("a-first").WithTimestamp(now).Build(),
						model_builders.NewConversationMessageBuilder().WithConversationId("b").WithContent("b-only").WithTimestamp(now).Build(),
					})

					assert.NoError(t, err)

					messages, err := suite.sut.ListByConversationId(context.Background(), "a")
					assert.NoError(t, err)
					assert.Len(t, messages, 2)
					assert.Equal(t, "a-first", lo.FromPtr(messages[0].Content))
					assert.Equal(t, "a-second", lo.FromPtr(messages[1].Content))

					otherMessages, err := suite.sut.ListByConversationId(context.Background(), "b")
					assert.NoError(t, err)
					assert.Len(t, otherMessages, 1)
					assert.Equal(t, "b-only", lo.FromPtr(otherMessages[0].Content))
				})
			})
		})
	}
}
