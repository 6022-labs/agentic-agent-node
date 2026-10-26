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
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
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

func WhenUpdatingMCPConnectionBeforeEach(tc *common_tests.DatabaseTestCase) *WhenUpdatingMCPConnectionTestingSuite {
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

	return &WhenUpdatingMCPConnectionTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenUpdatingMCPConnectionAfterEach(suite *WhenUpdatingMCPConnectionTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingMCPConnection(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a connection exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should overwrite the persisted fields", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingMCPConnectionBeforeEach(tc)
					defer WhenUpdatingMCPConnectionAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					inserted, err := suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("alpha").
							WithUrl("https://a.example").
							WithHeaderAuthKeys([]string{"X-Old"}).
							Build(),
					)
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
