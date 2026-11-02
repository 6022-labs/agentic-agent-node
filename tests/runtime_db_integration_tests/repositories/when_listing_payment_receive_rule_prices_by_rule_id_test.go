package repositories_test

import (
	"context"
	"testing"
	"time"

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

type WhenListingPaymentReceiveRulePricesByRuleIdTestingSuite struct {
	dbName                       string
	dbClient                     *runtime_db.DatabaseClient
	databaseContainerBuilder     container_builders.DatabaseContainerBuilderInterface
	sut                          *repositories.PaymentReceiveRulePriceRepository
	paymentReceiveRuleRepository *repositories.PaymentReceiveRuleRepository
	rule                         *models.PaymentReceiveRule
}

func WhenListingPaymentReceiveRulePricesByRuleIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingPaymentReceiveRulePricesByRuleIdTestingSuite {
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

	return &WhenListingPaymentReceiveRulePricesByRuleIdTestingSuite{
		dbName:                       dbName,
		dbClient:                     dbClient,
		databaseContainerBuilder:     databaseContainerBuilder,
		sut:                          repositories.NewPaymentReceiveRulePriceRepository(dbClient),
		paymentReceiveRuleRepository: paymentReceiveRuleRepository,
		rule:                         rule,
	}
}

func WhenListingPaymentReceiveRulePricesByRuleIdAfterEach(
	suite *WhenListingPaymentReceiveRulePricesByRuleIdTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingPaymentReceiveRulePricesByRuleId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no prices", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty list", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentReceiveRulePricesByRuleIdBeforeEach(tc)
					defer WhenListingPaymentReceiveRulePricesByRuleIdAfterEach(suite)

					prices, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Empty(t, prices)
				})
			})

			t.Run("Given the rule has several prices", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them oldest first", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentReceiveRulePricesByRuleIdBeforeEach(tc)
					defer WhenListingPaymentReceiveRulePricesByRuleIdAfterEach(suite)

					older := model_builders.NewPaymentReceiveRulePriceBuilder().
						WithPaymentReceiveRuleId(suite.rule.Id).
						WithPrice("2000").
						Build()
					older.CreatedAt = time.Now().Add(-time.Hour)
					newer := model_builders.NewPaymentReceiveRulePriceBuilder().
						WithPaymentReceiveRuleId(suite.rule.Id).
						WithPrice("0").
						Build()
					newer.CreatedAt = time.Now()
					_, err := suite.sut.Create(context.Background(), newer)
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(), older)
					assert.NoError(t, err)

					prices, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					if assert.Len(t, prices, 2) {
						assert.Equal(t, "2000", prices[0].Price)
						assert.Equal(t, "0", prices[1].Price)
					}
				})
			})

			t.Run("Given another rule has prices", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return only that rule's prices", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentReceiveRulePricesByRuleIdBeforeEach(tc)
					defer WhenListingPaymentReceiveRulePricesByRuleIdAfterEach(suite)

					other, err := suite.paymentReceiveRuleRepository.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.rule.PolicyId).Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().WithPaymentReceiveRuleId(other.Id).Build())
					assert.NoError(t, err)
					own, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentReceiveRulePriceBuilder().
							WithPaymentReceiveRuleId(suite.rule.Id).
							Build(),
					)
					assert.NoError(t, err)

					prices, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					if assert.Len(t, prices, 1) {
						assert.Equal(t, own.Id, prices[0].Id)
					}
				})
			})
		})
	}
}
