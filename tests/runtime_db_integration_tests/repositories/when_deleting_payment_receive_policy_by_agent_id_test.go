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

type WhenDeletingPaymentReceivePolicyByAgentIdTestingSuite struct {
	dbName                       string
	dbClient                     *runtime_db.DatabaseClient
	databaseContainerBuilder     container_builders.DatabaseContainerBuilderInterface
	sut                          *repositories.PaymentReceivePolicyRepository
	paymentReceiveRuleRepository *repositories.PaymentReceiveRuleRepository
}

func WhenDeletingPaymentReceivePolicyByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingPaymentReceivePolicyByAgentIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	return &WhenDeletingPaymentReceivePolicyByAgentIdTestingSuite{
		dbName:                       dbName,
		dbClient:                     dbClient,
		databaseContainerBuilder:     databaseContainerBuilder,
		sut:                          repositories.NewPaymentReceivePolicyRepository(dbClient),
		paymentReceiveRuleRepository: repositories.NewPaymentReceiveRuleRepository(dbClient),
	}
}

func WhenDeletingPaymentReceivePolicyByAgentIdAfterEach(suite *WhenDeletingPaymentReceivePolicyByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentReceivePolicyByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceivePolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentReceivePolicyByAgentIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), uuid.New()))
				})
			})

			t.Run("Given a policy with rules exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove the head", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceivePolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentReceivePolicyByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					policy, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)
					_, err = suite.paymentReceiveRuleRepository.Create(
						context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build(),
					)
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), policy.AgentId))

					gone, err := suite.sut.GetByAgentId(context.Background(), policy.AgentId)

					assert.NoError(t, err)
					assert.Nil(t, gone)
				})

				t.Run("Should remove its rules", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceivePolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentReceivePolicyByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					policy, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)
					_, err = suite.paymentReceiveRuleRepository.Create(
						context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build(),
					)
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), policy.AgentId))

					rules, err := suite.paymentReceiveRuleRepository.ListByPolicyId(context.Background(), policy.Id)

					assert.NoError(t, err)
					assert.Empty(t, rules)
				})

				t.Run("Should remove its rules' prices", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceivePolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentReceivePolicyByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					policy, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)
					rule, err := suite.paymentReceiveRuleRepository.Create(
						context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build(),
					)
					assert.NoError(t, err)
					paymentReceiveRulePriceRepository := repositories.NewPaymentReceiveRulePriceRepository(
						suite.dbClient,
					)
					_, err = paymentReceiveRulePriceRepository.Create(
						context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().WithPaymentReceiveRuleId(rule.Id).Build(),
					)
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), policy.AgentId))

					prices, err := paymentReceiveRulePriceRepository.ListByRuleId(context.Background(), rule.Id)

					assert.NoError(t, err)
					assert.Empty(t, prices)
				})
			})
		})
	}
}
