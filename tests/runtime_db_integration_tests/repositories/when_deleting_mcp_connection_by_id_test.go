package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenDeletingMCPConnectionByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.MCPConnectionRepository
}

func WhenDeletingMCPConnectionByIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingMCPConnectionByIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenDeletingMCPConnectionByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenDeletingMCPConnectionByIdAfterEach(suite *WhenDeletingMCPConnectionByIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingMCPConnectionById(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the connection does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingMCPConnectionByIdBeforeEach(tc)
					defer WhenDeletingMCPConnectionByIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), uuid.New(), uuid.New()))
				})
			})

			t.Run("Given a connection exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove the row", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingMCPConnectionByIdBeforeEach(tc)
					defer WhenDeletingMCPConnectionByIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(
						context.Background(),
						model_builders.NewMCPConnectionBuilder().
							WithAgentId(agentId).
							WithName("alpha").
							WithUrl("https://a.example").
							WithHeaderAuthKeys(nil).
							Build(),
					)
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), inserted.AgentId, inserted.Id))

					got, err := suite.sut.GetById(context.Background(), inserted.AgentId, inserted.Id)
					assert.NoError(t, err)
					assert.Nil(t, got)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.MCPConnection{}).Count(&count)
					assert.Equal(t, int64(0), count)
				})
			})
		})
	}
}
