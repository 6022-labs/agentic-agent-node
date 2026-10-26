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

type WhenListingMCPConnectionsByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.MCPConnectionRepository
}

func WhenListingMCPConnectionsByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingMCPConnectionsByAgentIdTestingSuite {
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

	return &WhenListingMCPConnectionsByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenListingMCPConnectionsByAgentIdAfterEach(suite *WhenListingMCPConnectionsByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingMCPConnectionsByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no connections exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty slice", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingMCPConnectionsByAgentIdBeforeEach(tc)
					defer WhenListingMCPConnectionsByAgentIdAfterEach(suite)

					list, err := suite.sut.ListByAgentId(context.Background(), uuid.New())
					assert.NoError(t, err)
					assert.Empty(t, list)
				})
			})

			t.Run("Given multiple connections exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them ordered by created_at ASC", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingMCPConnectionsByAgentIdBeforeEach(tc)
					defer WhenListingMCPConnectionsByAgentIdAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					alpha := model_builders.NewMCPConnectionBuilder().
						WithAgentId(agentId).
						WithName("alpha").
						WithUrl("https://a.example").
						WithHeaderAuthKeys(nil).
						WithCreatedAt(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)).
						Build()
					beta := model_builders.NewMCPConnectionBuilder().
						WithAgentId(agentId).
						WithName("beta").
						WithUrl("https://b.example").
						WithHeaderAuthKeys(nil).
						WithCreatedAt(time.Date(2026, 1, 1, 11, 0, 0, 0, time.UTC)).
						Build()

					_, err := suite.sut.Create(context.Background(), alpha)
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(), beta)
					assert.NoError(t, err)

					list, err := suite.sut.ListByAgentId(context.Background(), agentId)
					assert.NoError(t, err)
					assert.Len(t, list, 2)
					assert.Equal(t, "alpha", list[0].Name)
					assert.Equal(t, "beta", list[1].Name)
				})
			})
		})
	}
}
