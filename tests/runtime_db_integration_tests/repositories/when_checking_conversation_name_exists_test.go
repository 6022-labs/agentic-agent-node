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

type WhenCheckingConversationNameExistsTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenCheckingConversationNameExistsBeforeEach(
	tc *repositoryTestCase,
) *WhenCheckingConversationNameExistsTestingSuite {
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

	return &WhenCheckingConversationNameExistsTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenCheckingConversationNameExistsAfterEach(suite *WhenCheckingConversationNameExistsTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCheckingConversationNameExists(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the agent holds that name in the caller-less namespace", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it taken", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(ctx, agentId, nil, "Audit", nil)

					assert.NoError(t, err)
					assert.True(t, taken)
				})
			})

			t.Run("Given the holder itself is excluded in the caller-less namespace", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it free", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(
						ctx,
						agentId,
						nil,
						"Audit",
						lo.ToPtr("c1"),
					)

					assert.NoError(t, err)
					assert.False(t, taken)
				})
			})

			t.Run("Given a different conversation is excluded in the caller-less namespace", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it taken", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(
						ctx,
						agentId,
						nil,
						"Audit",
						lo.ToPtr("c2"),
					)

					assert.NoError(t, err)
					assert.True(t, taken)
				})
			})

			t.Run("Given the caller holds that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it taken", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(
						ctx,
						agentId,
						lo.ToPtr(alice),
						"Audit",
						nil,
					)

					assert.NoError(t, err)
					assert.True(t, taken)
				})
			})

			t.Run("Given another caller holds that name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it free", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					bob := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(ctx, agentId, lo.ToPtr(bob), "Audit", nil)

					assert.NoError(t, err)
					assert.False(t, taken)
				})
			})

			t.Run("Given a caller holds that name and the check is in the caller-less namespace", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it free", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(ctx, agentId, nil, "Audit", nil)

					assert.NoError(t, err)
					assert.False(t, taken)
				})
			})

			t.Run("Given a caller-less conversation holds that name and a caller checks it", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it free", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(
						ctx,
						agentId,
						lo.ToPtr(alice),
						"Audit",
						nil,
					)

					assert.NoError(t, err)
					assert.False(t, taken)
				})
			})

			t.Run("Given the caller's own holder is excluded", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it free", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithCallerId(alice).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(
						ctx,
						agentId,
						lo.ToPtr(alice),
						"Audit",
						lo.ToPtr("c1"),
					)

					assert.NoError(t, err)
					assert.False(t, taken)
				})
			})

			t.Run("Given the name is unused in the caller-less namespace", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it free", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(ctx, agentId, nil, "Never used", nil)

					assert.NoError(t, err)
					assert.False(t, taken)
				})
			})

			t.Run("Given another agent holds that name in the caller-less namespace", func(t *testing.T) {
				t.Parallel()

				t.Run("Should report it free", func(t *testing.T) {
					t.Parallel()
					suite := WhenCheckingConversationNameExistsBeforeEach(tc)
					defer WhenCheckingConversationNameExistsAfterEach(suite)

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

					taken, err := suite.sut.ExistsByAgentIdAndCallerIdAndName(ctx, secondAgentId, nil, "Audit", nil)

					assert.NoError(t, err)
					assert.False(t, taken)
				})
			})
		})
	}
}
