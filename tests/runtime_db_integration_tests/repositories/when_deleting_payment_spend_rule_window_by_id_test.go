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

type WhenDeletingPaymentSpendRuleWindowByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleWindowRepository
	rule                     *models.PaymentSpendRule
}

func WhenDeletingPaymentSpendRuleWindowByIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingPaymentSpendRuleWindowByIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentSpendPolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}
	rule, err := repositories.NewPaymentSpendRuleRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build())
	if err != nil {
		panic(err)
	}

	return &WhenDeletingPaymentSpendRuleWindowByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleWindowRepository(dbClient),
		rule:                     rule,
	}
}

func WhenDeletingPaymentSpendRuleWindowByIdAfterEach(suite *WhenDeletingPaymentSpendRuleWindowByIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentSpendRuleWindowById(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the window does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentSpendRuleWindowByIdBeforeEach(tc)
					defer WhenDeletingPaymentSpendRuleWindowByIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), uuid.New()))
				})
			})

			t.Run("Given several windows exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove only the targeted window", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentSpendRuleWindowByIdBeforeEach(tc)
					defer WhenDeletingPaymentSpendRuleWindowByIdAfterEach(suite)

					first, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithWindowSeconds(3600).
							Build())
					assert.NoError(t, err)
					second, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithWindowSeconds(86400).
							Build())
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), first.Id))

					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					if assert.Len(t, windows, 1) {
						assert.Equal(t, second.Id, windows[0].Id)
					}
				})
			})
		})
	}
}
