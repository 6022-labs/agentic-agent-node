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

type WhenGettingPaymentReceivePolicyByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceivePolicyRepository
}

func WhenGettingPaymentReceivePolicyByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentReceivePolicyByAgentIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenGettingPaymentReceivePolicyByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceivePolicyRepository(dbClient),
	}
}

func WhenGettingPaymentReceivePolicyByAgentIdAfterEach(suite *WhenGettingPaymentReceivePolicyByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentReceivePolicyByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentReceivePolicyByAgentIdBeforeEach(tc)
					defer WhenGettingPaymentReceivePolicyByAgentIdAfterEach(suite)

					policy, err := suite.sut.GetByAgentId(context.Background(), uuid.New())

					assert.NoError(t, err)
					assert.Nil(t, policy)
				})
			})

			t.Run("Given a policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentReceivePolicyByAgentIdBeforeEach(tc)
					defer WhenGettingPaymentReceivePolicyByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					created, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)

					policy, err := suite.sut.GetByAgentId(context.Background(), created.AgentId)

					assert.NoError(t, err)
					assert.Equal(t, created.Id, policy.Id)
					assert.Equal(t, created.AgentId, policy.AgentId)
				})
			})
		})
	}
}
