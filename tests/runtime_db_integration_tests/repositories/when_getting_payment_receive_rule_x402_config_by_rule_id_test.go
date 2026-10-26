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
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenGettingPaymentReceiveRuleX402ConfigByRuleIdTestingSuite struct {
	dbName                       string
	dbClient                     *runtime_db.DatabaseClient
	databaseContainerBuilder     container_builders.DatabaseContainerBuilderInterface
	sut                          *repositories.PaymentReceiveRuleX402ConfigRepository
	paymentReceiveRuleRepository *repositories.PaymentReceiveRuleRepository
	rule                         *models.PaymentReceiveRule
}

func WhenGettingPaymentReceiveRuleX402ConfigByRuleIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentReceiveRuleX402ConfigByRuleIdTestingSuite {
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
	if err := dbClient.GetDB(context.Background()).
		Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
		panic(err)
	}

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

	return &WhenGettingPaymentReceiveRuleX402ConfigByRuleIdTestingSuite{
		dbName:                       dbName,
		dbClient:                     dbClient,
		databaseContainerBuilder:     databaseContainerBuilder,
		sut:                          repositories.NewPaymentReceiveRuleX402ConfigRepository(dbClient),
		paymentReceiveRuleRepository: paymentReceiveRuleRepository,
		rule:                         rule,
	}
}

func WhenGettingPaymentReceiveRuleX402ConfigByRuleIdAfterEach(
	suite *WhenGettingPaymentReceiveRuleX402ConfigByRuleIdTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentReceiveRuleX402ConfigByRuleId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no config", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentReceiveRuleX402ConfigByRuleIdBeforeEach(tc)
					defer WhenGettingPaymentReceiveRuleX402ConfigByRuleIdAfterEach(suite)

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Nil(t, stored)
				})
			})

			t.Run("Given the rule is deleted", func(t *testing.T) {
				t.Parallel()

				t.Run("Should cascade-delete the config", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentReceiveRuleX402ConfigByRuleIdBeforeEach(tc)
					defer WhenGettingPaymentReceiveRuleX402ConfigByRuleIdAfterEach(suite)

					config := model_builders.NewPaymentReceiveRuleX402ConfigBuilder().
						WithPaymentReceiveRuleId(suite.rule.Id).
						Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					assert.NoError(
						t,
						suite.paymentReceiveRuleRepository.DeleteById(context.Background(), suite.rule.Id),
					)

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Nil(t, stored)
				})
			})
		})
	}
}
