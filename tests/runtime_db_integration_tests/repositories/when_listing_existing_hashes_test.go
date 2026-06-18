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

type WhenListingExistingHashesTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationMessageRepository
}

func WhenListingExistingHashesBeforeEach(tc *repositoryTestCase) *WhenListingExistingHashesTestingSuite {
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

	return &WhenListingExistingHashesTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationMessageRepository(dbClient),
	}
}

func WhenListingExistingHashesAfterEach(suite *WhenListingExistingHashesTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func newHashableMessage(conversationId, content string) models.ConversationMessage {
	return model_builders.NewConversationMessageBuilder().
		WithConversationId(conversationId).
		WithRole("user").
		WithContent(content).
		WithTimestamp(time.Now().UTC().Truncate(time.Microsecond)).
		Build()
}

func TestWhenListingExistingHashes(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given an empty input", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without touching the DB", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingExistingHashesBeforeEach(tc)
					defer WhenListingExistingHashesAfterEach(suite)

					existing, err := suite.sut.ListExistingHashes(context.Background(), nil)
					assert.NoError(t, err)
					assert.Empty(t, existing)
				})
			})

			t.Run("Given a partial match against stored hashes", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return only the subset present in the DB", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingExistingHashesBeforeEach(tc)
					defer WhenListingExistingHashesAfterEach(suite)

					stored := newHashableMessage("conv-partial", "stored")
					_, err := suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{stored})
					assert.NoError(t, err)

					unknown := newHashableMessage("conv-partial", "unknown")

					existing, err := suite.sut.ListExistingHashes(context.Background(), []string{stored.Hash, unknown.Hash})
					assert.NoError(t, err)
					assert.ElementsMatch(t, []string{stored.Hash}, existing)
				})
			})

			t.Run("Given every queried hash is already stored", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return all of them", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingExistingHashesBeforeEach(tc)
					defer WhenListingExistingHashesAfterEach(suite)

					a := newHashableMessage("conv-all-match", "alpha")
					b := newHashableMessage("conv-all-match", "beta")
					_, err := suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{a, b})
					assert.NoError(t, err)

					existing, err := suite.sut.ListExistingHashes(context.Background(), []string{a.Hash, b.Hash})
					assert.NoError(t, err)
					assert.ElementsMatch(t, []string{a.Hash, b.Hash}, existing)
				})
			})

			t.Run("Given none of the queried hashes are stored", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty slice", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingExistingHashesBeforeEach(tc)
					defer WhenListingExistingHashesAfterEach(suite)

					ghost1 := newHashableMessage("conv-none-match", "never-stored-1")
					ghost2 := newHashableMessage("conv-none-match", "never-stored-2")

					existing, err := suite.sut.ListExistingHashes(context.Background(), []string{ghost1.Hash, ghost2.Hash})
					assert.NoError(t, err)
					assert.Empty(t, existing)
				})
			})

			t.Run("Given the queried hashes belong to other conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should still return them — hashes are globally unique", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingExistingHashesBeforeEach(tc)
					defer WhenListingExistingHashesAfterEach(suite)

					otherConv := newHashableMessage("conv-other", "shared content")
					_, err := suite.sut.AppendMessages(context.Background(), []models.ConversationMessage{otherConv})
					assert.NoError(t, err)

					existing, err := suite.sut.ListExistingHashes(context.Background(), []string{otherConv.Hash})
					assert.NoError(t, err)
					assert.ElementsMatch(t, []string{otherConv.Hash}, existing)
				})
			})
		})
	}
}
