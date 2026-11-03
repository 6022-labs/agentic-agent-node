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

type WhenGettingPaymentSpendRuleByIdForUpdateTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	unitOfWorkFactory        *runtime_db_services.UnitOfWorkFactory
	sut                      *repositories.PaymentSpendRuleRepository
	rule                     *models.PaymentSpendRule
}

func WhenGettingPaymentSpendRuleByIdForUpdateBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentSpendRuleByIdForUpdateTestingSuite {
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

	return &WhenGettingPaymentSpendRuleByIdForUpdateTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		unitOfWorkFactory:        runtime_db_services.NewUnitOfWorkFactory(dbClient),
		sut:                      repositories.NewPaymentSpendRuleRepository(dbClient),
		rule:                     rule,
	}
}

func WhenGettingPaymentSpendRuleByIdForUpdateAfterEach(suite *WhenGettingPaymentSpendRuleByIdForUpdateTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentSpendRuleByIdForUpdate(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleByIdForUpdateBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleByIdForUpdateAfterEach(suite)

					unitOfWork, err := suite.unitOfWorkFactory.Start(context.Background())
					assert.NoError(t, err)
					defer func() { _ = unitOfWork.Rollback() }()

					rule, err := suite.sut.GetByIdForUpdate(unitOfWork.Context(), uuid.New())

					assert.NoError(t, err)
					assert.Nil(t, rule)
				})
			})

			t.Run("Given the rule exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it inside the scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleByIdForUpdateBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleByIdForUpdateAfterEach(suite)

					unitOfWork, err := suite.unitOfWorkFactory.Start(context.Background())
					assert.NoError(t, err)
					defer func() { _ = unitOfWork.Rollback() }()

					rule, err := suite.sut.GetByIdForUpdate(unitOfWork.Context(), suite.rule.Id)

					assert.NoError(t, err)
					if assert.NotNil(t, rule) {
						assert.Equal(t, suite.rule.Name, rule.Name)
					}
					assert.NoError(t, unitOfWork.Commit())
				})
			})
		})
	}
}
