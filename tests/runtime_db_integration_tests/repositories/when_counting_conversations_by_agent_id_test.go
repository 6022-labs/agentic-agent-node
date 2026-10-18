package repositories_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/samber/lo"
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

			t.Run("Given the caller owns the conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should count them", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingConversationsByAgentIdBeforeEach(tc)
					defer WhenCountingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					bob := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("alice-1").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("alice-2").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("bob-1").
						WithAgentId(agentId).
						WithCallerId(bob).
						Build())
					assert.NoError(t, err)

					total, err := suite.sut.CountByAgentId(ctx, agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					assert.Equal(t, uint(2), total)
				})
			})

			t.Run("Given another caller owns the conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return zero", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingConversationsByAgentIdBeforeEach(tc)
					defer WhenCountingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					bob := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("bob-1").
						WithAgentId(agentId).
						WithCallerId(bob).
						Build())
					assert.NoError(t, err)

					total, err := suite.sut.CountByAgentId(ctx, agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					assert.Equal(t, uint(0), total)
				})
			})

			t.Run("Given a caller-less conversation and a caller scope", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return zero", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingConversationsByAgentIdBeforeEach(tc)
					defer WhenCountingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("console").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)

					total, err := suite.sut.CountByAgentId(ctx, agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					assert.Equal(t, uint(0), total)
				})
			})

			t.Run("Given no caller scope", func(t *testing.T) {
				t.Parallel()

				t.Run("Should count every caller's conversations", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingConversationsByAgentIdBeforeEach(tc)
					defer WhenCountingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					bob := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("alice-1").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("bob-1").
						WithAgentId(agentId).
						WithCallerId(bob).
						Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("console").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)

					total, err := suite.sut.CountByAgentId(ctx, agentId, nil)

					assert.NoError(t, err)
					assert.Equal(t, uint(3), total)
				})
			})

			t.Run("Given the conversations belong to another agent", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return zero", func(t *testing.T) {
					t.Parallel()
					suite := WhenCountingConversationsByAgentIdBeforeEach(tc)
					defer WhenCountingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					otherAgentId := uuid.New()
					seedAgent(suite.dbClient, otherAgentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("theirs").
						WithAgentId(otherAgentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)

					total, err := suite.sut.CountByAgentId(ctx, agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					assert.Equal(t, uint(0), total)
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

					total, err := suite.sut.CountByAgentId(ctx, agentId, nil)

					assert.NoError(t, err)
					assert.Equal(t, uint(0), total)
				})
			})
		})
	}
}
