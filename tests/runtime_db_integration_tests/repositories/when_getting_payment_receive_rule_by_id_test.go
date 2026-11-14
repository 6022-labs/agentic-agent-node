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

type WhenGettingPaymentReceiveRuleByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceiveRuleRepository
	policy                   *models.PaymentReceivePolicy
}

func WhenGettingPaymentReceiveRuleByIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentReceiveRuleByIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentReceivePolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}

	return &WhenGettingPaymentReceiveRuleByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceiveRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenGettingPaymentReceiveRuleByIdAfterEach(suite *WhenGettingPaymentReceiveRuleByIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentReceiveRuleById(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentReceiveRuleByIdBeforeEach(tc)
					defer WhenGettingPaymentReceiveRuleByIdAfterEach(suite)

					rule, err := suite.sut.GetById(context.Background(), uuid.New())

					assert.NoError(t, err)
					assert.Nil(t, rule)
				})
			})

			t.Run("Given the rule exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it with the patterns decoded", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentReceiveRuleByIdBeforeEach(tc)
					defer WhenGettingPaymentReceiveRuleByIdAfterEach(suite)

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule, err := suite.sut.GetById(context.Background(), created.Id)

					assert.NoError(t, err)
					assert.Equal(t, created.Name, rule.Name)
					assert.Equal(t, created.Method, rule.Method)
				})
			})
		})
	}
}
