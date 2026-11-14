package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

type WhenUpsertingAgentTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentRepository
}

func WhenUpsertingAgentBeforeEach(tc *common_tests.DatabaseTestCase) *WhenUpsertingAgentTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenUpsertingAgentTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentRepository(dbClient),
	}
}

func WhenUpsertingAgentAfterEach(suite *WhenUpsertingAgentTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingAgent(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given an empty draft", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert with NULL identity columns", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					inserted, err := suite.sut.Upsert(context.Background(), &models.Agent{})

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)

					reloaded, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Nil(t, reloaded.ChainId)
					assert.Nil(t, reloaded.Name)
					assert.Nil(t, reloaded.EnsDomain)
					assert.Nil(t, reloaded.Creator)
				})
			})

			t.Run("Given no agent exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the new agent with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert"}).
						Build())

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)
				})

				t.Run("Should store a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert"}).
						Build())
					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.Agent{}).Count(&count)

					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given an agent already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should preserve its Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					first, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert"}).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Babbage").
						WithAttributes(map[string]string{"role": "facilitator"}).
						Build())

					assert.NoError(t, err)
					assert.Equal(t, first.Id, second.Id)
				})

				t.Run("Should replace its name", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert"}).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Babbage").
						WithAttributes(map[string]string{"role": "facilitator"}).
						Build())

					assert.NoError(t, err)
					assert.Equal(t, "Babbage", lo.FromPtr(second.Name))
				})

				t.Run("Should replace its attributes", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert"}).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Babbage").
						WithAttributes(map[string]string{"role": "facilitator"}).
						Build())

					assert.NoError(t, err)
					assert.Equal(t, "facilitator", second.Attributes["role"])
				})

				t.Run("Should replace its soul", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithSoul("old soul").
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithSoul("new soul").
						Build())
					assert.NoError(t, err)

					reloaded, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, "new soul", lo.FromPtr(reloaded.Soul))
				})

				t.Run("Should keep a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert"}).
						Build())
					assert.NoError(t, err)

					_, err = suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Babbage").
						WithAttributes(map[string]string{"role": "facilitator"}).
						Build())

					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.Agent{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
