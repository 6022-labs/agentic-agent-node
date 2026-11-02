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

type WhenDeletingPaymentReceiveRulePricesByRuleIdTestingSuite struct {
	dbName                       string
	dbClient                     *runtime_db.DatabaseClient
	databaseContainerBuilder     container_builders.DatabaseContainerBuilderInterface
	sut                          *repositories.PaymentReceiveRulePriceRepository
	paymentReceiveRuleRepository *repositories.PaymentReceiveRuleRepository
	rule                         *models.PaymentReceiveRule
}

func WhenDeletingPaymentReceiveRulePricesByRuleIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingPaymentReceiveRulePricesByRuleIdTestingSuite {
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

	return &WhenDeletingPaymentReceiveRulePricesByRuleIdTestingSuite{
		dbName:                       dbName,
		dbClient:                     dbClient,
		databaseContainerBuilder:     databaseContainerBuilder,
		sut:                          repositories.NewPaymentReceiveRulePriceRepository(dbClient),
		paymentReceiveRuleRepository: paymentReceiveRuleRepository,
		rule:                         rule,
	}
}

func WhenDeletingPaymentReceiveRulePricesByRuleIdAfterEach(
	suite *WhenDeletingPaymentReceiveRulePricesByRuleIdTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentReceiveRulePricesByRuleId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no prices", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceiveRulePricesByRuleIdBeforeEach(tc)
					defer WhenDeletingPaymentReceiveRulePricesByRuleIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteByRuleId(context.Background(), suite.rule.Id))
				})
			})

			t.Run("Given several rules have prices", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove only the targeted rule's prices", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentReceiveRulePricesByRuleIdBeforeEach(tc)
					defer WhenDeletingPaymentReceiveRulePricesByRuleIdAfterEach(suite)

					other, err := suite.paymentReceiveRuleRepository.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.rule.PolicyId).Build())
					assert.NoError(t, err)
					kept, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().WithPaymentReceiveRuleId(other.Id).Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPrice("0").
							Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							WithPrice("1000").
							Build())
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteByRuleId(context.Background(), suite.rule.Id))

					gone, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Empty(t, gone)
					others, err := suite.sut.ListByRuleId(context.Background(), other.Id)
					assert.NoError(t, err)
					if assert.Len(t, others, 1) {
						assert.Equal(t, kept.Id, others[0].Id)
					}
				})
			})
		})
	}
}
