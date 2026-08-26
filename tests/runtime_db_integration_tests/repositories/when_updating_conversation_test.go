package repositories_test

import (
	"context"
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

func WhenUpdatingConversationBeforeEach(tc *repositoryTestCase) *WhenUpdatingConversationTestingSuite {
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

	return &WhenUpdatingConversationTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenUpdatingConversationAfterEach(suite *WhenUpdatingConversationTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingConversation(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a stored conversation", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the new name", func(t *testing.T) {
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
					stored.Name = lo.ToPtr("Solidity audits")

					updated, err := suite.sut.Update(ctx, stored)

					assert.NoError(t, err)
					assert.Equal(t, "Solidity audits", *updated.Name)

					reloaded, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.Equal(t, "Solidity audits", *reloaded.Name)
				})

				t.Run("Should keep the creation timestamp and bump the update one", func(t *testing.T) {
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
					createdAt := stored.CreatedAt

					stored.Name = lo.ToPtr("Renamed")
					_, err = suite.sut.Update(ctx, stored)
					assert.NoError(t, err)

					reloaded, err := suite.sut.GetById(ctx, "c1")
					assert.NoError(t, err)
					assert.WithinDuration(t, createdAt, reloaded.CreatedAt, time.Second)
					assert.False(t, reloaded.UpdatedAt.Before(reloaded.CreatedAt))
				})

				t.Run("Should clear the name when it is set back to nil", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingConversationBeforeEach(tc)
					defer WhenUpdatingConversationAfterEach(suite)

					ctx := context.Background()
					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)
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
		})
	}
}
