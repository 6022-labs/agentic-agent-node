package repositories_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type WhenGettingConversationByIdAndAgentIdAndCallerIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingConversationByIdAndAgentIdAndCallerIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenGettingConversationByIdAndAgentIdAndCallerIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(
	suite *WhenGettingConversationByIdAndAgentIdAndCallerIdTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingConversationByIdAndAgentIdAndCallerId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the caller owns the conversation", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(tc)
					defer WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)

					conversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "c1", agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					if assert.NotNil(t, conversation) {
						assert.Equal(t, "c1", conversation.Id)
					}
				})
			})

			t.Run("Given another caller owns the conversation", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(tc)
					defer WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					bob := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithCallerId(bob).
						Build())
					assert.NoError(t, err)

					conversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "c1", agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					assert.Nil(t, conversation)
				})
			})

			t.Run("Given a caller-less conversation and a caller scope", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(tc)
					defer WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)

					conversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "c1", agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					assert.Nil(t, conversation)
				})
			})

			t.Run("Given no caller scope", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return any caller's conversation", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(tc)
					defer WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
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

					aliceConversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "alice-1", agentId, nil)
					assert.NoError(t, err)
					bobConversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "bob-1", agentId, nil)
					assert.NoError(t, err)

					if assert.NotNil(t, aliceConversation) {
						assert.Equal(t, "alice-1", aliceConversation.Id)
					}
					if assert.NotNil(t, bobConversation) {
						assert.Equal(t, "bob-1", bobConversation.Id)
					}
				})

				t.Run("Should return a caller-less conversation", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(tc)
					defer WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
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

					conversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "console", agentId, nil)

					assert.NoError(t, err)
					if assert.NotNil(t, conversation) {
						assert.Equal(t, "console", conversation.Id)
					}
				})
			})

			t.Run("Given the conversation belongs to another agent", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(tc)
					defer WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					otherAgentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, otherAgentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(otherAgentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)

					conversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "c1", agentId, lo.ToPtr(alice))

					assert.NoError(t, err)
					assert.Nil(t, conversation)
				})
			})

			t.Run("Given the conversation does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdAndAgentIdAndCallerIdBeforeEach(tc)
					defer WhenGettingConversationByIdAndAgentIdAndCallerIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					conversation, err := suite.sut.GetByIdAndAgentIdAndCallerId(ctx, "ghost", agentId, nil)

					assert.NoError(t, err)
					assert.Nil(t, conversation)
				})
			})
		})
	}
}
