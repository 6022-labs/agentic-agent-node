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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenUpdatingSelfToolTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.SelfToolRepository
}

func WhenUpdatingSelfToolBeforeEach(tc *repositoryTestCase) *WhenUpdatingSelfToolTestingSuite {
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

	return &WhenUpdatingSelfToolTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenUpdatingSelfToolAfterEach(suite *WhenUpdatingSelfToolTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingSelfTool(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the tool does not exist in database", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert the row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingSelfToolBeforeEach(tc)
					defer WhenUpdatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					ghost := model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("ghost").WithEnabled(true).Build()

					_, err := suite.sut.Update(context.Background(), ghost)
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), agentId, ghost.Id)
					assert.NoError(t, err)
					assert.NotNil(t, got)
				})
			})

			t.Run("Given an existing tool", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the toggled Enabled flag and bump UpdatedAt", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingSelfToolBeforeEach(tc)
					defer WhenUpdatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(context.Background(), model_builders.NewSelfToolBuilder().WithAgentId(agentId).WithName("bash").WithEnabled(true).Build())
					assert.NoError(t, err)
					originalUpdatedAt := inserted.UpdatedAt

					time.Sleep(50 * time.Millisecond)

					inserted.Enabled = false
					updated, err := suite.sut.Update(context.Background(), inserted)
					assert.NoError(t, err)
					assert.False(t, updated.Enabled)

					got, err := suite.sut.GetById(context.Background(), inserted.AgentId, inserted.Id)
					assert.NoError(t, err)
					assert.NotNil(t, got)
					assert.False(t, got.Enabled)
					assert.True(t, got.UpdatedAt.After(originalUpdatedAt))
				})
			})
		})
	}
}
