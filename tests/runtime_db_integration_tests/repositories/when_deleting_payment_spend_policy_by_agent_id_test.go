package repositories_test

import (
	"context"
	"testing"

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

type WhenDeletingPaymentSpendPolicyByAgentIdTestingSuite struct {
	dbName                     string
	dbClient                   *runtime_db.DatabaseClient
	databaseContainerBuilder   container_builders.DatabaseContainerBuilderInterface
	sut                        *repositories.PaymentSpendPolicyRepository
	paymentSpendRuleRepository *repositories.PaymentSpendRuleRepository
}

func WhenDeletingPaymentSpendPolicyByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingPaymentSpendPolicyByAgentIdTestingSuite {
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

	return &WhenDeletingPaymentSpendPolicyByAgentIdTestingSuite{
		dbName:                     dbName,
		dbClient:                   dbClient,
		databaseContainerBuilder:   databaseContainerBuilder,
		sut:                        repositories.NewPaymentSpendPolicyRepository(dbClient),
		paymentSpendRuleRepository: repositories.NewPaymentSpendRuleRepository(dbClient),
	}
}

func WhenDeletingPaymentSpendPolicyByAgentIdAfterEach(suite *WhenDeletingPaymentSpendPolicyByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentSpendPolicyByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentSpendPolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentSpendPolicyByAgentIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), uuid.New()))
				})
			})

			t.Run("Given a policy with rules exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove the head and its rules", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentSpendPolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentSpendPolicyByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					policy, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)
					rule, err := suite.paymentSpendRuleRepository.Create(
						context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build(),
					)
					assert.NoError(t, err)
					configRepository := repositories.NewPaymentSpendRuleX402ConfigRepository(suite.dbClient)
					assert.NoError(t, configRepository.Upsert(context.Background(),
						model_builders.NewPaymentSpendRuleX402ConfigBuilder().
							WithPaymentSpendRuleId(rule.Id).
							WithPolicyId(policy.Id).
							Build()))
					windowRepository := repositories.NewPaymentSpendRuleWindowRepository(suite.dbClient)
					_, err = windowRepository.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(rule.Id).Build())
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), policy.AgentId))

					gone, err := suite.sut.GetByAgentId(context.Background(), policy.AgentId)
					assert.NoError(t, err)
					assert.Nil(t, gone)

					rules, err := suite.paymentSpendRuleRepository.ListByPolicyId(context.Background(), policy.Id)
					assert.NoError(t, err)
					assert.Empty(t, rules)

					// The config and window rows ride the rule delete via the DB cascade.
					config, err := configRepository.GetByRuleId(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Nil(t, config)

					windows, err := windowRepository.ListByRuleId(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Empty(t, windows)
				})
			})
		})
	}
}
