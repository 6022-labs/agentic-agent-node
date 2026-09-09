package repositories_test

import (
	"context"
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

				t.Run("Should store every pattern array", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentSpendRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentSpendRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
						WithPaymentSpendRuleId(suite.rule.Id).
						WithNetwork("eip155:80002").
						WithAssetTransferMethod(enums.AssetTransferMethodEip3009).
						Build()

					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, "eip155:80002", stored.Network)
					assert.Equal(t, enums.AssetTransferMethodEip3009, *stored.AssetTransferMethod)
					assert.Equal(t, config.Asset, stored.Asset)
					assert.Equal(t, config.PayToPatterns, stored.PayToPatterns)
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
							WithPayToPatterns([]string{".*"}).
							Build()))

					assert.NoError(t, suite.sut.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithPayToPatterns([]string{"0x1111111111111111111111111111111111111111"}).
							Build()))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, []string{"0x1111111111111111111111111111111111111111"}, stored.PayToPatterns)

					var count int64
					assert.NoError(t, suite.dbClient.GetDB(context.Background()).
						Model(&models.PaymentSpendRuleX402Config{}).
						Where("payment_spend_rule_id = ?", suite.rule.Id).
						Count(&count).Error)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
