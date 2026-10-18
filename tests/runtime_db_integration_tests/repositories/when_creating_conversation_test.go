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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenCreatingConversationTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenCreatingConversationBeforeEach(tc *repositoryTestCase) *WhenCreatingConversationTestingSuite {
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

	return &WhenCreatingConversationTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenCreatingConversationAfterEach(suite *WhenCreatingConversationTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingConversation(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a free name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the head", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					created, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithName("ERC20 mint audit").
						Build())

					assert.NoError(t, err)
					assert.Equal(t, "c1", created.Id)

					stored, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.Equal(t, "ERC20 mint audit", *stored.Name)
					assert.Equal(t, agentId, stored.AgentId)
				})
			})

			t.Run("Given no name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist a nil name", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.Nil(t, stored.Name)
				})
			})

			t.Run("Given a caller id", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the caller id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())

					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					if assert.NotNil(t, stored) && assert.NotNil(t, stored.CallerId) {
						assert.Equal(t, alice, *stored.CallerId)
					}
				})
			})

			t.Run("Given the caller holds that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should reject the insert", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())

					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("first").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("second").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())

					assertUniqueViolation(t, err)
				})
			})

			t.Run("Given several unnamed heads", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert them all", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					// NULL names stay distinct under a unique index on every driver.
					for _, id := range []string{"unnamed-1", "unnamed-2"} {
						_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
							WithId(id).
							WithAgentId(agentId).
							Build())
						assert.NoError(t, err)
					}
				})
			})

			t.Run("Given another caller holds that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should accept the insert", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					bob := strings.ToLower(common_tests.RandomEthAddress().Hex())

					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("first").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("second").
						WithAgentId(agentId).
						WithCallerId(bob).
						WithName("Audit").
						Build())

					assert.NoError(t, err)
				})
			})

			t.Run("Given a caller-less conversation uses that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should accept the insert", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())

					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("first").
						WithAgentId(agentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("second").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())

					assert.NoError(t, err)
				})
			})

			t.Run("Given two caller-less conversations share a name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should accept the insert", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("first").
						WithAgentId(agentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					// NULL caller ids stay distinct under the unique index; the use case guards this namespace.
					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("second").
						WithAgentId(agentId).
						WithName("Audit").
						Build())

					assert.NoError(t, err)
				})
			})

			t.Run("Given another agent holds that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should accept the insert", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingConversationBeforeEach(tc)
					defer WhenCreatingConversationAfterEach(suite)

					ctx := context.Background()
					firstAgentId := uuid.New()
					secondAgentId := uuid.New()
					seedAgent(suite.dbClient, firstAgentId)
					seedAgent(suite.dbClient, secondAgentId)

					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(firstAgentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c2").
						WithAgentId(secondAgentId).
						WithName("Audit").
						Build())

					assert.NoError(t, err)
				})
			})
		})
	}
}
