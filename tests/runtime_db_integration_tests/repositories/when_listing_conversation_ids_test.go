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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenListingConversationIdsTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationMessageRepository
}

func WhenListingConversationIdsBeforeEach(tc *repositoryTestCase) *WhenListingConversationIdsTestingSuite {
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

	return &WhenListingConversationIdsTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationMessageRepository(dbClient),
	}
}

func WhenListingConversationIdsAfterEach(suite *WhenListingConversationIdsTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingConversationIds(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no messages were ever appended", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty slice", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationIdsBeforeEach(tc)
					defer WhenListingConversationIdsAfterEach(suite)

					ids, err := suite.sut.ListConversationIds(context.Background())

					assert.NoError(t, err)
					assert.Empty(t, ids)
				})
			})

			t.Run("Given two conversations with different last-activity", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return distinct ids ordered by most-recent activity first", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationIdsBeforeEach(tc)
					defer WhenListingConversationIdsAfterEach(suite)

					now := time.Now().UTC().Truncate(time.Microsecond)
					older := model_builders.NewConversationMessageBuilder().
						WithConversationId("old").WithContent("x").WithTimestamp(now.Add(-time.Hour)).Build()
					newer := model_builders.NewConversationMessageBuilder().
						WithConversationId("new").WithContent("y").WithTimestamp(now).Build()
					newerSecond := model_builders.NewConversationMessageBuilder().
						WithConversationId("new").WithContent("z").WithTimestamp(now.Add(time.Minute)).Build()

					_, err := suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{older, newer, newerSecond})
					assert.NoError(t, err)

					ids, err := suite.sut.ListConversationIds(context.Background())

					assert.NoError(t, err)
					assert.Equal(t, []string{"new", "old"}, ids)
				})
			})
		})
	}
}
