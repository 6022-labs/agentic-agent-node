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

type WhenListingPaymentSpendRulesByPolicyIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleRepository
	policy                   *models.PaymentSpendPolicy
}

func WhenListingPaymentSpendRulesByPolicyIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingPaymentSpendRulesByPolicyIdTestingSuite {
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

	return &WhenListingPaymentSpendRulesByPolicyIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenListingPaymentSpendRulesByPolicyIdAfterEach(suite *WhenListingPaymentSpendRulesByPolicyIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingPaymentSpendRulesByPolicyId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty list", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentSpendRulesByPolicyIdBeforeEach(tc)
					defer WhenListingPaymentSpendRulesByPolicyIdAfterEach(suite)

					rules, err := suite.sut.ListByPolicyId(context.Background(), suite.policy.Id)

					assert.NoError(t, err)
					assert.Empty(t, rules)
				})
			})

			t.Run("Given several rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them in creation order", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentSpendRulesByPolicyIdBeforeEach(tc)
					defer WhenListingPaymentSpendRulesByPolicyIdAfterEach(suite)

					_, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().
							WithPolicyId(suite.policy.Id).
							WithName("first").
							Build(),
					)
					assert.NoError(t, err)
					_, err = suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().
							WithPolicyId(suite.policy.Id).
							WithName("second").
							Build(),
					)
					assert.NoError(t, err)

					rules, err := suite.sut.ListByPolicyId(context.Background(), suite.policy.Id)

					assert.NoError(t, err)
					if assert.Len(t, rules, 2) {
						assert.Equal(t, "first", rules[0].Name)
						assert.Equal(t, "second", rules[1].Name)
					}
				})
			})
		})
	}
}
