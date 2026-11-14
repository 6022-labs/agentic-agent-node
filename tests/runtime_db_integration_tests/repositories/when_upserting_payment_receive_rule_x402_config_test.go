package repositories_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenUpsertingPaymentReceiveRuleX402ConfigTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceiveRuleX402ConfigRepository
	rule                     *models.PaymentReceiveRule
}

func WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpsertingPaymentReceiveRuleX402ConfigTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentReceivePolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}
	rule, err := repositories.NewPaymentReceiveRuleRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(policy.Id).Build())
	if err != nil {
		panic(err)
	}

	return &WhenUpsertingPaymentReceiveRuleX402ConfigTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceiveRuleX402ConfigRepository(dbClient),
		rule:                     rule,
	}
}

func WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite *WhenUpsertingPaymentReceiveRuleX402ConfigTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingPaymentReceiveRuleX402Config(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no config yet", func(t *testing.T) {
				t.Parallel()

				t.Run("Should read back its network", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
						WithPaymentReceiveRuleId(suite.rule.Id).
						WithPolicyId(suite.rule.PolicyId).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Equal(t, config.Network, stored.Network)
				})

				t.Run("Should read back its policy and token", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
						WithPaymentReceiveRuleId(suite.rule.Id).
						WithPolicyId(suite.rule.PolicyId).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Equal(t, suite.rule.PolicyId, stored.PolicyId)
					assert.Equal(t, config.Asset, stored.Asset)
					assert.Equal(t, config.PayTo, stored.PayTo)
				})

				t.Run("Should read back the EIP-712 domain", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
						WithPaymentReceiveRuleId(suite.rule.Id).
						WithPolicyId(suite.rule.PolicyId).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Equal(t, config.Eip712Name, stored.Eip712Name)
					assert.Equal(t, config.Eip712Version, stored.Eip712Version)
				})
			})

			t.Run("Given a permit2 config with no EIP-712 domain", func(t *testing.T) {
				t.Parallel()

				t.Run("Should round-trip the domain as NULL", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
						WithPaymentReceiveRuleId(suite.rule.Id).
						WithPolicyId(suite.rule.PolicyId).
						WithAssetTransferMethod(enums.AssetTransferMethodPermit2).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Equal(t, enums.AssetTransferMethodPermit2, stored.AssetTransferMethod)
					assert.Nil(t, stored.Eip712Name)
					assert.Nil(t, stored.Eip712Version)
				})
			})

			t.Run("Given a config already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace it on upsert", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithNetwork("eip155:80002").
							Build()))
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithNetwork("eip155:8453").
							Build()))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Equal(t, "eip155:8453", stored.Network)
				})

				t.Run("Should keep a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							Build()))
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							Build()))

					var count int64
					assert.NoError(t, suite.dbClient.GetDB(context.Background()).
						Model(&models.PaymentReceiveRuleX402Config{}).
						Where("payment_receive_rule_id = ?", suite.rule.Id).
						Count(&count).Error)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given another rule already prices the token", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse a second rule for that token", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					other, err := repositories.NewPaymentReceiveRuleRepository(suite.dbClient).Create(
						context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.rule.PolicyId).Build(),
					)
					assert.NoError(t, err)
					pricedAsset := strings.ToLower(common_tests.RandomEthAddress().Hex())
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(other.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(pricedAsset).
							Build()))

					err = suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(pricedAsset).
							Build())

					assert.Error(t, err)
					stored, getErr := suite.sut.GetByRuleId(context.Background(), other.Id)
					assert.NoError(t, getErr)
					assert.Equal(t, other.Id, stored.PaymentReceiveRuleId)
				})
			})

			// The rule already has a config, so the upsert takes its update path rather than an insert.
			t.Run("Given a rule moves to a token another rule already prices", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse it and leave the other rule untouched", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentReceiveRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentReceiveRuleX402ConfigAfterEach(suite)

					other, err := repositories.NewPaymentReceiveRuleRepository(suite.dbClient).Create(
						context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.rule.PolicyId).Build(),
					)
					assert.NoError(t, err)
					pricedAsset := strings.ToLower(common_tests.RandomEthAddress().Hex())
					otherPayTo := common_tests.RandomEthAddress().Hex()
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(other.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(pricedAsset).
							WithPayTo(otherPayTo).
							Build()))
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							Build()))

					err = suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(pricedAsset).
							Build())

					assert.Error(t, err)
					stored, getErr := suite.sut.GetByRuleId(context.Background(), other.Id)
					assert.NoError(t, getErr)
					assert.Equal(t, pricedAsset, stored.Asset)
					assert.Equal(t, otherPayTo, stored.PayTo)
				})
			})
		})
	}
}
