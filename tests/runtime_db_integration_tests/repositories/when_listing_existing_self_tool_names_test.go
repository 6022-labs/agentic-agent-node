package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenListingExistingSelfToolNamesTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.SelfToolRepository
}

func WhenListingExistingSelfToolNamesBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingExistingSelfToolNamesTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenListingExistingSelfToolNamesTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewSelfToolRepository(dbClient),
	}
}

func WhenListingExistingSelfToolNamesAfterEach(suite *WhenListingExistingSelfToolNamesTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingExistingSelfToolNames(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given some of the given names have rows", func(t *testing.T) {
				t.Parallel()

				t.Run("Should list only the existing ones for that agent", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingExistingSelfToolNamesBeforeEach(tc)
					defer WhenListingExistingSelfToolNamesAfterEach(suite)

					agentId := uuid.New()
					otherAgentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					seeds.SeedAgent(suite.dbClient, otherAgentId)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(agentId).
							WithName("bash").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewSelfToolBuilder().
							WithAgentId(otherAgentId).
							WithName("curl").
							WithEnabled(true).
							Build(),
					)
					assert.NoError(t, err)

					existing, err := suite.sut.ListExistingNames(
						context.Background(),
						agentId,
						[]string{"bash", "curl", "unknown"},
					)
					assert.NoError(t, err)
					assert.Equal(t, []string{"bash"}, existing)
				})
			})

			t.Run("Given no names are provided", func(t *testing.T) {
				t.Parallel()

				t.Run("Should list nothing without querying", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingExistingSelfToolNamesBeforeEach(tc)
					defer WhenListingExistingSelfToolNamesAfterEach(suite)

					existing, err := suite.sut.ListExistingNames(context.Background(), uuid.New(), nil)
					assert.NoError(t, err)
					assert.Empty(t, existing)
				})
			})
		})
	}
}
