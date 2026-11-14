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

type WhenDeletingPaymentReceiveRuleByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceiveRuleRepository
	policy                   *models.PaymentReceivePolicy
}

func WhenDeletingPaymentReceiveRuleByIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingPaymentReceiveRuleByIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentReceivePolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}

	return &WhenDeletingPaymentReceiveRuleByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceiveRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenDeletingPaymentReceiveRuleByIdAfterEach(suite *WhenDeletingPaymentReceiveRuleByIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentReceiveRuleById(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceiveRuleByIdBeforeEach(tc)
					defer WhenDeletingPaymentReceiveRuleByIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), uuid.New()))
				})
			})

			t.Run("Given several rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove only the targeted rule", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceiveRuleByIdBeforeEach(tc)
					defer WhenDeletingPaymentReceiveRuleByIdAfterEach(suite)

					first, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)
					second, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), first.Id))

					rules, err := suite.sut.ListByPolicyId(context.Background(), suite.policy.Id)
					assert.NoError(t, err)
					if assert.Len(t, rules, 1) {
						assert.Equal(t, second.Id, rules[0].Id)
					}
				})
			})

			t.Run("Given the rule has prices", func(t *testing.T) {
				t.Parallel()

				t.Run("Should delete its prices too", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceiveRuleByIdBeforeEach(tc)
					defer WhenDeletingPaymentReceiveRuleByIdAfterEach(suite)

					rule, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)
					paymentReceiveRulePriceRepository := repositories.NewPaymentReceiveRulePriceRepository(
						suite.dbClient,
					)
					_, err = paymentReceiveRulePriceRepository.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().WithPaymentReceiveRuleId(rule.Id).Build())
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), rule.Id))

					prices, err := paymentReceiveRulePriceRepository.ListByRuleId(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Empty(t, prices)
				})
			})
		})
	}
}
