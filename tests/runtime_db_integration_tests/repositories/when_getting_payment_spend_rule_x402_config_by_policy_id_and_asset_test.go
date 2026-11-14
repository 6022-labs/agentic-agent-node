package repositories_test

import (
	"context"
	"strings"
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

type WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleX402ConfigRepository
	rule                     *models.PaymentSpendRule
}

func WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetTestingSuite {
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

	return &WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleX402ConfigRepository(dbClient),
		rule:                     rule,
	}
}

func WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetAfterEach(
	suite *WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAsset(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no rule names the asset", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetAfterEach(suite)
					asset := strings.ToLower(common_tests.RandomEthAddress().Hex())

					config, err := suite.sut.GetByPolicyIdAndAsset(
						context.Background(),
						suite.rule.PolicyId,
						"eip155:80002",
						asset,
					)

					assert.NoError(t, err)
					assert.Nil(t, config)
				})
			})

			t.Run("Given another policy names the asset", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetAfterEach(suite)
					asset := strings.ToLower(common_tests.RandomEthAddress().Hex())

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(asset).
							Build()))

					config, err := suite.sut.GetByPolicyIdAndAsset(
						context.Background(),
						uuid.New(),
						"eip155:80002",
						asset,
					)

					assert.NoError(t, err)
					assert.Nil(t, config)
				})
			})

			t.Run("Given a rule of the policy names the asset", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return its config", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByPolicyIdAndAssetAfterEach(suite)
					asset := strings.ToLower(common_tests.RandomEthAddress().Hex())

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithNetwork("eip155:80002").
							WithAsset(asset).
							Build()))

					config, err := suite.sut.GetByPolicyIdAndAsset(
						context.Background(),
						suite.rule.PolicyId,
						"eip155:80002",
						asset,
					)

					assert.NoError(t, err)
					if assert.NotNil(t, config) {
						assert.Equal(t, suite.rule.Id, config.PaymentSpendRuleId)
					}
				})
			})
		})
	}
}
