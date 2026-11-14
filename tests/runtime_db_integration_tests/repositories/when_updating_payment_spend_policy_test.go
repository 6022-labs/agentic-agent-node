package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenUpdatingPaymentSpendPolicyTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendPolicyRepository
	policy                   *models.PaymentSpendPolicy
}

func WhenUpdatingPaymentSpendPolicyBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpdatingPaymentSpendPolicyTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentSpendPolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}

	return &WhenUpdatingPaymentSpendPolicyTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendPolicyRepository(dbClient),
		policy:                   policy,
	}
}

func WhenUpdatingPaymentSpendPolicyAfterEach(suite *WhenUpdatingPaymentSpendPolicyTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace its payees", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentSpendPolicyBeforeEach(tc)
					defer WhenUpdatingPaymentSpendPolicyAfterEach(suite)

					suite.policy.PayToPatterns = []string{"0xabc.*"}
					_, err := suite.sut.Update(context.Background(), suite.policy)
					assert.NoError(t, err)

					stored, err := suite.sut.GetByAgentId(context.Background(), suite.policy.AgentId)
					assert.NoError(t, err)
					assert.Equal(t, suite.policy.Id, stored.Id)
					assert.Equal(t, []string{"0xabc.*"}, stored.PayToPatterns)
				})
			})

			t.Run("Given the payees are cleared", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store an empty payee list", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentSpendPolicyBeforeEach(tc)
					defer WhenUpdatingPaymentSpendPolicyAfterEach(suite)

					suite.policy.PayToPatterns = []string{}
					_, err := suite.sut.Update(context.Background(), suite.policy)
					assert.NoError(t, err)

					stored, err := suite.sut.GetByAgentId(context.Background(), suite.policy.AgentId)
					assert.NoError(t, err)
					assert.Equal(t, []string{}, stored.PayToPatterns)
				})
			})
		})
	}
}
