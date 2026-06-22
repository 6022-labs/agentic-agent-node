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

type WhenGettingPaymentRuleX402ConfigByRuleIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentRuleX402ConfigRepository
	paymentRuleRepository    *repositories.PaymentRuleRepository
	rule                     *models.PaymentRule
}

func WhenGettingPaymentRuleX402ConfigByRuleIdBeforeEach(tc *repositoryTestCase) *WhenGettingPaymentRuleX402ConfigByRuleIdTestingSuite {
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
	paymentRuleRepository := repositories.NewPaymentRuleRepository(dbClient)
	rule, err := paymentRuleRepository.Create(
		context.Background(), model_builders.NewPaymentRuleBuilder().WithPolicyId(policy.Id).Build())
	if err != nil {
		panic(err)
	}

	return &WhenGettingPaymentRuleX402ConfigByRuleIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentRuleX402ConfigRepository(dbClient),
		paymentRuleRepository:    paymentRuleRepository,
		rule:                     rule,
	}
}

func WhenGettingPaymentRuleX402ConfigByRuleIdAfterEach(suite *WhenGettingPaymentRuleX402ConfigByRuleIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentRuleX402ConfigByRuleId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no config", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentRuleX402ConfigByRuleIdBeforeEach(tc)
					defer WhenGettingPaymentRuleX402ConfigByRuleIdAfterEach(suite)

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Nil(t, stored)
				})
			})

			t.Run("Given the rule is deleted", func(t *testing.T) {
				t.Parallel()

				t.Run("Should cascade-delete the config", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentRuleX402ConfigByRuleIdBeforeEach(tc)
					defer WhenGettingPaymentRuleX402ConfigByRuleIdAfterEach(suite)

					config := model_builders.NewPaymentRuleX402ConfigBuilder().WithPaymentRuleId(suite.rule.Id).Build()
					assert.NoError(t, suite.sut.Upsert(context.Background(), config))

					assert.NoError(t, suite.paymentRuleRepository.DeleteById(context.Background(), suite.rule.Id))

					stored, err := suite.sut.GetByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Nil(t, stored)
				})
			})
		})
	}
}
