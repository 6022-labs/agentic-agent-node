package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenGettingAgentGuidanceByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentGuidanceRepository
}

func WhenGettingAgentGuidanceByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingAgentGuidanceByAgentIdTestingSuite {
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

	return &WhenGettingAgentGuidanceByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentGuidanceRepository(dbClient),
	}
}

func WhenGettingAgentGuidanceByAgentIdAfterEach(suite *WhenGettingAgentGuidanceByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingAgentGuidanceByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no guidance row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentGuidanceByAgentIdBeforeEach(tc)
					defer WhenGettingAgentGuidanceByAgentIdAfterEach(suite)

					got, err := suite.sut.GetByAgentId(context.Background(), uuid.New())
					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})

			t.Run("Given a guidance row exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the persisted row", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentGuidanceByAgentIdBeforeEach(tc)
					defer WhenGettingAgentGuidanceByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentGuidanceBuilder().
						WithAgentId(agentId).
						WithContent("answer concisely").
						Build())

					assert.NoError(t, err)

					got, err := suite.sut.GetByAgentId(context.Background(), agentId)
					assert.NoError(t, err)
					assert.Equal(t, inserted.Id, got.Id)
					assert.Equal(t, "answer concisely", got.Content)
				})
			})
		})
	}
}
