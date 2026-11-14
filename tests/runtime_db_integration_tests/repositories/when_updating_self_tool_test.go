package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenUpdatingSelfToolTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.SelfToolRepository
}

func WhenUpdatingSelfToolBeforeEach(tc *common_tests.DatabaseTestCase) *WhenUpdatingSelfToolTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenUpdatingSelfToolTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenUpdatingSelfToolAfterEach(suite *WhenUpdatingSelfToolTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingSelfTool(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the tool does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert the row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingSelfToolBeforeEach(tc)
					defer WhenUpdatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					ghost := model_builders.NewSelfToolBuilder().
						WithAgentId(agentId).
						WithName("ghost").
						WithEnabled(true).
						Build()

					_, err := suite.sut.Update(context.Background(), ghost)
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), agentId, ghost.Id)
					assert.NoError(t, err)
					assert.NotNil(t, got)
				})
			})

			t.Run("Given a tool exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the toggled Enabled flag", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingSelfToolBeforeEach(tc)
					defer WhenUpdatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					inserted.Enabled = false
					updated, err := suite.sut.Update(context.Background(), inserted)
					assert.NoError(t, err)
					assert.False(t, updated.Enabled)

					got, err := suite.sut.GetById(context.Background(), inserted.AgentId, inserted.Id)

					assert.NoError(t, err)
					if assert.NotNil(t, got) {
						assert.False(t, got.Enabled)
					}
				})

				t.Run("Should bump UpdatedAt", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingSelfToolBeforeEach(tc)
					defer WhenUpdatingSelfToolAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					inserted, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					originalUpdatedAt := inserted.UpdatedAt

					time.Sleep(50 * time.Millisecond)

					inserted.Enabled = false
					_, err = suite.sut.Update(context.Background(), inserted)
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), inserted.AgentId, inserted.Id)

					assert.NoError(t, err)
					if assert.NotNil(t, got) {
						assert.True(t, got.UpdatedAt.After(originalUpdatedAt))
					}
				})
			})
		})
	}
}
