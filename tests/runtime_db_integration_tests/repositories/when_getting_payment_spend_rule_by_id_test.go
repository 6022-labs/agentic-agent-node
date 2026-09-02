package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"

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

type WhenGettingPaymentSpendRuleByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleRepository
	policy                   *models.PaymentSpendPolicy
}

func WhenGettingPaymentSpendRuleByIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentSpendRuleByIdTestingSuite {
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

	return &WhenGettingPaymentSpendRuleByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenGettingPaymentSpendRuleByIdAfterEach(suite *WhenGettingPaymentSpendRuleByIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentSpendRuleById(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a cap at the full width of the column", func(t *testing.T) {
				t.Parallel()

				t.Run("Should round-trip a max-uint256 cap without truncating", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleByIdBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleByIdAfterEach(suite)

					maxUint256 := "115792089237316195423570985008687907853269984665640564039457584007913129639935"
					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().
							WithPolicyId(suite.policy.Id).
							WithMaxAmount(maxUint256).
							Build())
					assert.NoError(t, err)

					rule, err := suite.sut.GetById(context.Background(), created.Id)
					assert.NoError(t, err)
					assert.Equal(t, maxUint256, rule.MaxAmount)
				})
			})

			t.Run("Given the rule does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleByIdBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleByIdAfterEach(suite)

					rule, err := suite.sut.GetById(context.Background(), uuid.New())

					assert.NoError(t, err)
					assert.Nil(t, rule)
				})
			})

			t.Run("Given the rule exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it with its cap intact", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendRuleByIdBeforeEach(tc)
					defer WhenGettingPaymentSpendRuleByIdAfterEach(suite)

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule, err := suite.sut.GetById(context.Background(), created.Id)

					assert.NoError(t, err)
					assert.Equal(t, created.Name, rule.Name)
					assert.Equal(t, created.Method, rule.Method)
					assert.Equal(t, created.MaxAmount, rule.MaxAmount)
				})
			})
		})
	}
}
