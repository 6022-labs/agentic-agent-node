package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenUpdatingMCPConnectionTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.MCPConnectionRepository
}

func WhenUpdatingMCPConnectionBeforeEach(tc *repositoryTestCase) *WhenUpdatingMCPConnectionTestingSuite {
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

	return &WhenUpdatingMCPConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenUpdatingMCPConnectionAfterEach(suite *WhenUpdatingMCPConnectionTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingMCPConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given an existing connection", func(t *testing.T) {
				t.Parallel()

				t.Run("Should overwrite the persisted fields", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingMCPConnectionBeforeEach(tc)
					defer WhenUpdatingMCPConnectionAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(context.Background(), newMCPConnection(agentId, "alpha", "https://a.example", "X-Old"))
					assert.NoError(t, err)

					inserted.Name = "alpha-renamed"
					inserted.Url = "https://renamed.example"
					inserted.HeaderAuthKeys = models.MCPConnectionHeaderAuthKeys{"X-New"}
					updated, err := suite.sut.Update(context.Background(), inserted)
					assert.NoError(t, err)
					assert.Equal(t, "alpha-renamed", updated.Name)

					got, err := suite.sut.GetById(context.Background(), inserted.AgentId, inserted.Id)
					assert.NoError(t, err)
					assert.Equal(t, "alpha-renamed", got.Name)
					assert.Equal(t, "https://renamed.example", got.Url)
					assert.Len(t, got.HeaderAuthKeys, 1)
					assert.Equal(t, "X-New", string(got.HeaderAuthKeys[0]))
				})
			})
		})
	}
}
