package repositories_test

import (
	"context"
	"strings"
	"testing"

	"time"

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

type WhenUpdatingConversationTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenUpdatingConversationBeforeEach(tc *common_tests.DatabaseTestCase) *WhenUpdatingConversationTestingSuite {
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

	return &WhenUpdatingConversationTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenUpdatingConversationAfterEach(suite *WhenUpdatingConversationTestingSuite) {
	_ = suite.dbClient.Close()
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingConversation(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a stored conversation", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the new name", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingConversationBeforeEach(tc)
					defer WhenUpdatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					stored.Name = lo.ToPtr("Solidity audits")

					updated, err := suite.sut.Update(ctx, stored)

					assert.NoError(t, err)
					assert.Equal(t, "Solidity audits", *updated.Name)

					reloaded, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.Equal(t, "Solidity audits", *reloaded.Name)
				})

				t.Run("Should keep the creation timestamp", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingConversationBeforeEach(tc)
					defer WhenUpdatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						Build())
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					createdAt := stored.CreatedAt

					stored.Name = lo.ToPtr("Renamed")
					_, err = suite.sut.Update(ctx, stored)
					assert.NoError(t, err)

					reloaded, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.WithinDuration(t, createdAt, reloaded.CreatedAt, time.Second)
				})

				t.Run("Should bump the update timestamp", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingConversationBeforeEach(tc)
					defer WhenUpdatingConversationAfterEach(suite)

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
					stored.Name = lo.ToPtr("Renamed")
					_, err = suite.sut.Update(ctx, stored)
					assert.NoError(t, err)

					reloaded, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.False(t, reloaded.UpdatedAt.Before(reloaded.CreatedAt))
				})
			})

			t.Run("Given the name is set back to nil", func(t *testing.T) {
				t.Parallel()

				t.Run("Should clear the name", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingConversationBeforeEach(tc)
					defer WhenUpdatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}
					_, err := suite.sut.Create(ctx, model_builders.NewConversationBuilder().
						WithId("c1").
						WithAgentId(agentId).
						WithName("Audit").
						Build())
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					stored.Name = nil
					_, err = suite.sut.Update(ctx, stored)
					assert.NoError(t, err)

					reloaded, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.Nil(t, reloaded.Name)
				})
			})

			t.Run("Given a caller-scoped conversation is renamed", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep its caller id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingConversationBeforeEach(tc)
					defer WhenUpdatingConversationAfterEach(suite)

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
					stored.Name = lo.ToPtr("Renamed")
					_, err = suite.sut.Update(ctx, stored)
					assert.NoError(t, err)

					reloaded, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					if assert.NotNil(t, reloaded) && assert.NotNil(t, reloaded.CallerId) {
						assert.Equal(t, alice, *reloaded.CallerId)
					}
				})
			})
		})
	}
}
