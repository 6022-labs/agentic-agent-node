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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenUpsertingPaymentRuleX402ConfigTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentRuleX402ConfigRepository
	rule                     *models.PaymentRule
}

func WhenUpsertingPaymentRuleX402ConfigBeforeEach(tc *repositoryTestCase) *WhenUpsertingPaymentRuleX402ConfigTestingSuite {
	databaseContainerBuilder := tc.clientBuilderFn()
	dbName, connectionString := databaseContainerBuilder.Create()

	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	dbClient, err := runtime_db.NewDatabaseClient(
		&settings.DatabaseSettings{Driver: tc.driver, Url: connectionString},
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
	seedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentPolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentPolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}
	rule, err := repositories.NewPaymentRuleRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentRuleBuilder().WithPolicyId(policy.Id).Build())
	if err != nil {
		panic(err)
	}

	return &WhenUpsertingPaymentRuleX402ConfigTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentRuleX402ConfigRepository(dbClient),
		rule:                     rule,
	}
}

func WhenUpsertingPaymentRuleX402ConfigAfterEach(suite *WhenUpsertingPaymentRuleX402ConfigTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingPaymentRuleX402Config(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no config yet", func(t *testing.T) {
				t.Parallel()

				t.Run("Should upsert and read it back with the patterns decoded", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentRuleX402ConfigBuilder().
						WithPaymentRuleId(suite.rule.Id).
						WithAddressPatterns([]string{".*"}).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Equal(t, config.Network, stored.Network)
					assert.Equal(t, []string{".*"}, stored.AddressPatterns)
					// eip3009 keeps the token's EIP-712 domain.
					assert.Equal(t, config.Eip712Name, stored.Eip712Name)
					assert.Equal(t, config.Eip712Version, stored.Eip712Version)
				})
			})

			t.Run("Given a permit2 config with no EIP-712 domain", func(t *testing.T) {
				t.Parallel()

				t.Run("Should round-trip the domain as NULL", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentRuleX402ConfigAfterEach(suite)

					config := model_builders.NewPaymentRuleX402ConfigBuilder().
						WithPaymentRuleId(suite.rule.Id).
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

			t.Run("Given an existing config", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace it on upsert", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingPaymentRuleX402ConfigBeforeEach(tc)
					defer WhenUpsertingPaymentRuleX402ConfigAfterEach(suite)

					first := model_builders.NewPaymentRuleX402ConfigBuilder().WithPaymentRuleId(suite.rule.Id).Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), first))

					second := model_builders.NewPaymentRuleX402ConfigBuilder().
						WithPaymentRuleId(suite.rule.Id).
						WithAddressPatterns([]string{"0x1111111111111111111111111111111111111111"}).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), second))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Equal(t, []string{"0x1111111111111111111111111111111111111111"}, stored.AddressPatterns)
				})
			})
		})
	}
}
