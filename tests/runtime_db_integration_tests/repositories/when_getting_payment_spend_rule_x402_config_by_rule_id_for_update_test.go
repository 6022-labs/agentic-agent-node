package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	runtime_db_services "github.com/6022-labs/agentic-agent-node/src/runtime_db/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	unitOfWorkFactory        *runtime_db_services.UnitOfWorkFactory
	sut                      *repositories.PaymentSpendRuleX402ConfigRepository
	rule                     *models.PaymentSpendRule
	config                   *models.PaymentSpendRuleX402Config
}

func WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateTestingSuite {
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
	config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
		WithPaymentSpendRuleId(rule.Id).
		WithPolicyId(policy.Id).
		Build()
	if err := repositories.NewPaymentSpendRuleX402ConfigRepository(dbClient).
		Upsert(context.Background(), config); err != nil {
		panic(err)
	}

	return &WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		unitOfWorkFactory:        runtime_db_services.NewUnitOfWorkFactory(dbClient),
		sut:                      repositories.NewPaymentSpendRuleX402ConfigRepository(dbClient),
		rule:                     rule,
		config:                   config,
	}
}

func WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateAfterEach(
	suite *WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdate(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no token", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateAfterEach(suite)

					unitOfWork, err := suite.unitOfWorkFactory.Start(context.Background())
					assert.NoError(t, err)
					defer func() { _ = unitOfWork.Rollback() }()

					config, err := suite.sut.GetByRuleIdForUpdate(unitOfWork.Context(), uuid.New())

					assert.NoError(t, err)
					assert.Nil(t, config)
				})
			})

			t.Run("Given the rule has a token", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it inside the scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByRuleIdForUpdateAfterEach(suite)

					unitOfWork, err := suite.unitOfWorkFactory.Start(context.Background())
					assert.NoError(t, err)
					defer func() { _ = unitOfWork.Rollback() }()

					config, err := suite.sut.GetByRuleIdForUpdate(unitOfWork.Context(), suite.rule.Id)

					assert.NoError(t, err)
					if assert.NotNil(t, config) {
						assert.Equal(t, suite.config.Asset, config.Asset)
						assert.Equal(t, suite.config.Network, config.Network)
					}
					assert.NoError(t, unitOfWork.Commit())
				})
			})
		})
	}
}
