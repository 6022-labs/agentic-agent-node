package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
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

type WhenCreatingPaymentSpendRuleTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleRepository
	policy                   *models.PaymentSpendPolicy
}

func WhenCreatingPaymentSpendRuleBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenCreatingPaymentSpendRuleTestingSuite {
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

	return &WhenCreatingPaymentSpendRuleTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenCreatingPaymentSpendRuleAfterEach(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given rules are created one after another", func(t *testing.T) {
				t.Parallel()

				t.Run("Should not write the association, so the policy row is untouched", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendRuleBeforeEach(tc)
					defer WhenCreatingPaymentSpendRuleAfterEach(suite)

					// Create omits the association, or gorm would upsert the embedded policy as a second row.
					rule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(suite.policy.Id).Build()
					rule.Policy = *model_builders.NewPaymentSpendPolicyBuilder().Build()

					_, err := suite.sut.Create(context.Background(), rule)
					assert.NoError(t, err)

					var policies int64
					assert.NoError(t, suite.dbClient.GetDB(context.Background()).
						Model(&models.PaymentSpendPolicy{}).Count(&policies).Error)
					assert.Equal(t, int64(1), policies)
				})

				t.Run("Should generate an id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendRuleBeforeEach(tc)
					defer WhenCreatingPaymentSpendRuleAfterEach(suite)

					created, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().
							WithId(uuid.Nil).
							WithPolicyId(suite.policy.Id).
							Build(),
					)

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, created.Id)
				})
			})

			t.Run("Given a rule without a per-payment cap", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store the cap as unset", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendRuleBeforeEach(tc)
					defer WhenCreatingPaymentSpendRuleAfterEach(suite)

					created, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().
							WithPolicyId(suite.policy.Id).
							WithoutMaxAmountPerPayment().
							Build(),
					)
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(context.Background(), created.Id)
					assert.NoError(t, err)
					assert.Nil(t, stored.MaxAmountPerPayment)
				})
			})
		})
	}
}
