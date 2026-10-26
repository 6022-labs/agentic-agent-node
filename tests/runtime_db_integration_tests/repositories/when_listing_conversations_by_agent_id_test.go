package repositories_test

import (
	"context"
	"strings"
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
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenListingConversationsByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenListingConversationsByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingConversationsByAgentIdTestingSuite {
	databaseContainerBuilder := tc.ClientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := runtime_db.NewDatabaseClient(
		&settings.DatabaseSettings{Driver: tc.Driver, Url: connectionString},
		&settings.MigrationsSettings{FolderPath: *migrationFolderPath},
		zap.NewNop(),
	)
	if err != nil {
		panic(err)
	}
	if err := dbClient.Migrate(); err != nil {
		panic(err)
	}

	return &WhenListingConversationsByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenListingConversationsByAgentIdAfterEach(suite *WhenListingConversationsByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingConversationsByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the caller owns the conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should list them", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

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

					list, err := suite.sut.ListByAgentId(ctx, agentId, lo.ToPtr(alice), nil, nil)

					assert.NoError(t, err)
					assert.ElementsMatch(t, []string{"alice-1", "alice-2"}, lo.Map(
						list,
						func(conversation models.Conversation, _ int) string { return conversation.Id },
					))
				})
			})

			t.Run("Given another caller owns the conversations", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

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

					list, err := suite.sut.ListByAgentId(ctx, agentId, lo.ToPtr(alice), nil, nil)

					assert.NoError(t, err)
					assert.Empty(t, list)
				})
			})

			t.Run("Given a caller-less conversation and a caller scope", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("console").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)

					list, err := suite.sut.ListByAgentId(ctx, agentId, lo.ToPtr(alice), nil, nil)

					assert.NoError(t, err)
					assert.Empty(t, list)
				})
			})

			t.Run("Given no caller scope", func(t *testing.T) {
				t.Parallel()

				t.Run("Should list every caller's conversations", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

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

					list, err := suite.sut.ListByAgentId(ctx, agentId, nil, nil, nil)

					assert.NoError(t, err)
					assert.ElementsMatch(t, []string{"alice-1", "bob-1", "console"}, lo.Map(
						list,
						func(conversation models.Conversation, _ int) string { return conversation.Id },
					))
				})
			})

			t.Run("Given the conversations belong to another agent", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nothing", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

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

					list, err := suite.sut.ListByAgentId(ctx, agentId, lo.ToPtr(alice), nil, nil)

					assert.NoError(t, err)
					assert.Empty(t, list)
				})
			})

			t.Run("Given conversations with distinct creation times", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the newest first", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					createdAt := time.Now().UTC().Truncate(time.Second)
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("older").
						WithAgentId(agentId).
						WithCreatedAt(createdAt.Add(-time.Hour)).
						Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("newer").
						WithAgentId(agentId).
						WithCreatedAt(createdAt).
						Build())
					assert.NoError(t, err)

					list, err := suite.sut.ListByAgentId(ctx, agentId, nil, nil, nil)

					assert.NoError(t, err)
					assert.Equal(t, []string{"newer", "older"}, []string{list[0].Id, list[1].Id})
				})
			})

			t.Run("Given three conversations", func(t *testing.T) {
				t.Parallel()

				seed := func(suite *WhenListingConversationsByAgentIdTestingSuite, agentId uuid.UUID) {
					createdAt := time.Now().UTC().Truncate(time.Second)
					// Distinct creation times so the newest-first order is deterministic across drivers.
					for index, id := range []string{"a", "b", "c"} {
						_, err := suite.sut.Create(context.Background(), model_builders.NewConversationBuilder().
							WithId(id).
							WithAgentId(agentId).
							WithCreatedAt(createdAt.Add(-time.Duration(index)*time.Minute)).
							Build())
						assert.NoError(t, err)
					}
				}

				t.Run("Should return the first page", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					seed(suite, agentId)

					firstPage, err := suite.sut.ListByAgentId(ctx, agentId, nil, lo.ToPtr(0), lo.ToPtr(2))

					assert.NoError(t, err)
					if assert.Len(t, firstPage, 2) {
						assert.Equal(t, []string{"a", "b"}, []string{firstPage[0].Id, firstPage[1].Id})
					}
				})

				t.Run("Should return the remaining rows on the second page", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingConversationsByAgentIdBeforeEach(tc)
					defer WhenListingConversationsByAgentIdAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					seed(suite, agentId)

					secondPage, err := suite.sut.ListByAgentId(ctx, agentId, nil, lo.ToPtr(2), lo.ToPtr(2))

					assert.NoError(t, err)
					if assert.Len(t, secondPage, 1) {
						assert.Equal(t, "c", secondPage[0].Id)
					}
				})

				t.Run("Given an offset past the end", func(t *testing.T) {
					t.Parallel()

					t.Run("Should return nothing", func(t *testing.T) {
						t.Parallel()
						suite := WhenListingConversationsByAgentIdBeforeEach(tc)
						defer WhenListingConversationsByAgentIdAfterEach(suite)

						ctx := context.Background()
						agentId := uuid.New()
						seedAgent(suite.dbClient, agentId)
						seed(suite, agentId)

						list, err := suite.sut.ListByAgentId(ctx, agentId, nil, lo.ToPtr(99), lo.ToPtr(2))

						assert.NoError(t, err)
						assert.Empty(t, list)
					})
				})

				t.Run("Given only a limit", func(t *testing.T) {
					t.Parallel()

					t.Run("Should cap them", func(t *testing.T) {
						t.Parallel()
						suite := WhenListingConversationsByAgentIdBeforeEach(tc)
						defer WhenListingConversationsByAgentIdAfterEach(suite)

						ctx := context.Background()
						agentId := uuid.New()
						seedAgent(suite.dbClient, agentId)
						seed(suite, agentId)

						list, err := suite.sut.ListByAgentId(ctx, agentId, nil, nil, lo.ToPtr(1))

						assert.NoError(t, err)
						assert.Len(t, list, 1)
						assert.Equal(t, "a", list[0].Id)
					})
				})
			})
		})
	}
}
