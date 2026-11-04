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

type WhenDeletingPaymentSpendRuleByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleRepository
	policy                   *models.PaymentSpendPolicy
}

func WhenDeletingPaymentSpendRuleByIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenDeletingPaymentSpendRuleByIdTestingSuite {
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

	return &WhenDeletingPaymentSpendRuleByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenDeletingPaymentSpendRuleByIdAfterEach(suite *WhenDeletingPaymentSpendRuleByIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentSpendRuleById(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentSpendRuleByIdBeforeEach(tc)
					defer WhenDeletingPaymentSpendRuleByIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), uuid.New()))
				})
			})

			t.Run("Given several rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove only the targeted rule", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentSpendRuleByIdBeforeEach(tc)
					defer WhenDeletingPaymentSpendRuleByIdAfterEach(suite)

					first, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)
					second, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), first.Id))

					rules, err := suite.sut.ListByPolicyId(context.Background(), suite.policy.Id)
					assert.NoError(t, err)
					if assert.Len(t, rules, 1) {
						assert.Equal(t, second.Id, rules[0].Id)
					}
				})
			})
		})
	}
}
