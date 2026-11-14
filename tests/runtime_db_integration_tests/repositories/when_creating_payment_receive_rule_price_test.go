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

type WhenCreatingPaymentReceiveRulePriceTestingSuite struct {
	dbName                       string
	dbClient                     *runtime_db.DatabaseClient
	databaseContainerBuilder     container_builders.DatabaseContainerBuilderInterface
	sut                          *repositories.PaymentReceiveRulePriceRepository
	paymentReceiveRuleRepository *repositories.PaymentReceiveRuleRepository
	rule                         *models.PaymentReceiveRule
}

func WhenCreatingPaymentReceiveRulePriceBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenCreatingPaymentReceiveRulePriceTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentReceivePolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}
	paymentReceiveRuleRepository := repositories.NewPaymentReceiveRuleRepository(dbClient)
	rule, err := paymentReceiveRuleRepository.Create(
		context.Background(), model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build())
	if err != nil {
		panic(err)
	}

	return &WhenCreatingPaymentReceiveRulePriceTestingSuite{
		dbName:                       dbName,
		dbClient:                     dbClient,
		databaseContainerBuilder:     databaseContainerBuilder,
		sut:                          repositories.NewPaymentReceiveRulePriceRepository(dbClient),
		paymentReceiveRuleRepository: paymentReceiveRuleRepository,
		rule:                         rule,
	}
}

func WhenCreatingPaymentReceiveRulePriceAfterEach(suite *WhenCreatingPaymentReceiveRulePriceTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentReceiveRulePrice(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a price for a rule", func(t *testing.T) {
				t.Parallel()

				t.Run("Should generate an id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentReceiveRulePriceBeforeEach(tc)
					defer WhenCreatingPaymentReceiveRulePriceAfterEach(suite)

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithId(uuid.Nil).
							WithPaymentReceiveRuleId(suite.rule.Id).
							Build())
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, created.Id)
				})

				t.Run("Should read back its address patterns", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentReceiveRulePriceBeforeEach(tc)
					defer WhenCreatingPaymentReceiveRulePriceAfterEach(suite)

					_, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithAddressPatterns([]string{".*"}).
							Build())
					assert.NoError(t, err)

					prices, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					if assert.Len(t, prices, 1) {
						assert.Equal(t, []string{".*"}, prices[0].AddressPatterns)
					}
				})
			})

			t.Run("Given the rule already has that price", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse the second price", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentReceiveRulePriceBeforeEach(tc)
					defer WhenCreatingPaymentReceiveRulePriceAfterEach(suite)

					_, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPrice("1000").
							Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPrice("1000").
							Build())

					// Drivers word the unique violation differently, so the proof is that nothing was stored.
					assert.Error(t, err)
					prices, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Len(t, prices, 1)
				})
			})

			t.Run("Given another rule already has that price", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store it", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentReceiveRulePriceBeforeEach(tc)
					defer WhenCreatingPaymentReceiveRulePriceAfterEach(suite)

					other, err := suite.paymentReceiveRuleRepository.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.rule.PolicyId).Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(other.Id).
							WithPrice("1000").
							Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPrice("1000").
							Build())

					assert.NoError(t, err)
				})
			})
		})
	}
}
