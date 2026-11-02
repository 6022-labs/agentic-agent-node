package repositories_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenUpsertingPaymentSpendRuleX402ConfigTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleX402ConfigRepository
	rule                     *models.PaymentSpendRule
}

func WhenUpsertingPaymentSpendRuleX402ConfigBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpsertingPaymentSpendRuleX402ConfigTestingSuite {
	databaseContainerBuilder := tc.ClientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := runtime_db.NewDatabaseClient(
		&settings.DatabaseSettings{Driver: tc.Driver, Url: connectionString},
		&settings.MigrationsSettings{FolderPath: *migrationFolderPath},
		zap.NewNop(),
	)
	if err != nil {
		panic(err)
	}
	if err := dbClient.Migrate(); err != nil {
		panic(err)
	}

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

	return &WhenUpsertingPaymentSpendRuleX402ConfigTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleX402ConfigRepository(dbClient),
		rule:                     rule,
	}
}

func WhenUpsertingPaymentSpendRuleX402ConfigAfterEach(suite *WhenUpsertingPaymentSpendRuleX402ConfigTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingPaymentSpendRuleX402Config(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a config for a rule that has none", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store the token and its transfer method", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentSpendRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentSpendRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
						WithPaymentSpendRuleId(suite.rule.Id).
						WithPolicyId(suite.rule.PolicyId).
						WithNetwork("eip155:80002").
						WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
						Build()

					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, "eip155:80002", stored.Network)
					assert.Equal(t, enums.AssetTransferMethodEip3009, *stored.AssetTransferMethod)
					assert.Equal(t, config.Asset, stored.Asset)
					assert.Equal(t, suite.rule.PolicyId, stored.PolicyId)
				})
			})

			t.Run("Given a config is upserted twice", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace the row rather than add a second", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentSpendRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentSpendRuleX402ConfigAfterEach(suite)

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithNetwork("eip155:80002").
							Build()))

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithNetwork("eip155:8453").
							Build()))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, "eip155:8453", stored.Network)

					var count int64
					assert.NoError(t, suite.dbClient.GetDB(context.Background()).
						Model(&models.PaymentSpendRuleX402Config{}).
						Where("payment_spend_rule_id = ?", suite.rule.Id).
						Count(&count).Error)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given the rule's transfer method is cleared", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store it unset", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentSpendRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentSpendRuleX402ConfigAfterEach(suite)

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
							Build()))

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							Build()))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Nil(t, stored.AssetTransferMethod)
				})
			})

			t.Run("Given another rule of the policy already names the asset", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse a second rule for that token", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentSpendRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentSpendRuleX402ConfigAfterEach(suite)

					other, err := repositories.NewPaymentSpendRuleRepository(suite.dbClient).Create(
						context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(suite.rule.PolicyId).Build(),
					)
					assert.NoError(t, err)
					spentAsset := strings.ToLower(common_tests.RandomEthAddress().Hex())
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(other.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(spentAsset).
							Build()))

					err = suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(spentAsset).
							Build())

					assert.Error(t, err)
					stored, getErr := suite.sut.GetByRuleId(context.Background(), other.Id)
					assert.NoError(t, getErr)
					assert.Equal(t, other.Id, stored.PaymentSpendRuleId)
				})
			})

			// The rule already has a config, so the upsert takes its update path rather than an insert.
			t.Run("Given a configured rule moves onto a token another rule spends", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse it and leave the other rule's row unchanged", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentSpendRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentSpendRuleX402ConfigAfterEach(suite)

					other, err := repositories.NewPaymentSpendRuleRepository(suite.dbClient).Create(
						context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(suite.rule.PolicyId).Build(),
					)
					assert.NoError(t, err)
					spentAsset := strings.ToLower(common_tests.RandomEthAddress().Hex())
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(other.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(spentAsset).
							WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
							Build()))
					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(strings.ToLower(common_tests.RandomEthAddress().Hex())).
							Build()))

					err = suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPolicyId(suite.rule.PolicyId).
							WithAsset(spentAsset).
							Build())

					assert.Error(t, err)
					stored, getErr := suite.sut.GetByRuleId(context.Background(), other.Id)
					assert.NoError(t, getErr)
					assert.Equal(t, spentAsset, stored.Asset)
					assert.Equal(t, enums.AssetTransferMethodEip3009, *stored.AssetTransferMethod)
				})
			})
		})
	}
}
