package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	runtime_common_repositories "github.com/6022-labs/agentic-agent-node/src/runtime_common/repositories"
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

type WhenGettingPaymentSpendRuleX402ConfigByRuleIdTestingSuite struct {
	dbName                     string
	dbClient                   *runtime_db.DatabaseClient
	databaseContainerBuilder   container_builders.DatabaseContainerBuilderInterface
	sut                        *repositories.PaymentSpendRuleX402ConfigRepository
	unitOfWorkFactory          *runtime_db_services.UnitOfWorkFactory
	paymentSpendRuleRepository *repositories.PaymentSpendRuleRepository
	rule                       *models.PaymentSpendRule
}

func WhenGettingPaymentSpendRuleX402ConfigByRuleIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentSpendRuleX402ConfigByRuleIdTestingSuite {
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
	paymentSpendRuleRepository := repositories.NewPaymentSpendRuleRepository(dbClient)
	rule, err := paymentSpendRuleRepository.Create(
		context.Background(), model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build())
	if err != nil {
		panic(err)
	}

	return &WhenGettingPaymentSpendRuleX402ConfigByRuleIdTestingSuite{
		dbName:                     dbName,
		dbClient:                   dbClient,
		databaseContainerBuilder:   databaseContainerBuilder,
		sut:                        repositories.NewPaymentSpendRuleX402ConfigRepository(dbClient),
		unitOfWorkFactory:          runtime_db_services.NewUnitOfWorkFactory(dbClient),
		paymentSpendRuleRepository: paymentSpendRuleRepository,
		rule:                       rule,
	}
}

func WhenGettingPaymentSpendRuleX402ConfigByRuleIdAfterEach(
	suite *WhenGettingPaymentSpendRuleX402ConfigByRuleIdTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentSpendRuleX402ConfigByRuleId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no config", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByRuleIdBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByRuleIdAfterEach(suite)

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Nil(t, stored)
				})
			})

			t.Run("Given the rule is deleted", func(t *testing.T) {
				t.Parallel()

				t.Run("Should cascade-delete the config", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByRuleIdBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByRuleIdAfterEach(suite)

					config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
						WithPaymentSpendRuleId(suite.rule.Id).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					assert.NoError(
						t,
						suite.paymentSpendRuleRepository.DeleteById(context.Background(), suite.rule.Id),
					)

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Nil(t, stored)
				})
			})

			t.Run("Given the config is read to lock inside a unit of work", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleX402ConfigByRuleIdBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleX402ConfigByRuleIdAfterEach(suite)

					config := model_builders.NewPaymentSpendRuleX402ConfigBuilder().
						WithPaymentSpendRuleId(suite.rule.Id).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))
					unitOfWork, err := suite.unitOfWorkFactory.Start(context.Background())
					assert.NoError(t, err)
					defer func() { _ = unitOfWork.Rollback() }()

					stored, err := suite.sut.GetByRuleId(
						unitOfWork.Context(),
						suite.rule.Id,
						runtime_common_repositories.ToLock,
					)

					assert.NoError(t, err)
					if assert.NotNil(t, stored) {
						assert.Equal(t, config.Asset, stored.Asset)
					}
				})
			})
		})
	}
}
