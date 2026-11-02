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

type WhenListingPaymentSpendRuleWindowsByRuleIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleWindowRepository
	rule                     *models.PaymentSpendRule
}

func WhenListingPaymentSpendRuleWindowsByRuleIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingPaymentSpendRuleWindowsByRuleIdTestingSuite {
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

	return &WhenListingPaymentSpendRuleWindowsByRuleIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleWindowRepository(dbClient),
		rule:                     rule,
	}
}

func WhenListingPaymentSpendRuleWindowsByRuleIdAfterEach(
	suite *WhenListingPaymentSpendRuleWindowsByRuleIdTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingPaymentSpendRuleWindowsByRuleId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule has no windows", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty list", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentSpendRuleWindowsByRuleIdBeforeEach(tc)
					defer WhenListingPaymentSpendRuleWindowsByRuleIdAfterEach(suite)

					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Empty(t, windows)
				})
			})

			t.Run("Given the rule has several windows", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them shortest period first", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentSpendRuleWindowsByRuleIdBeforeEach(tc)
					defer WhenListingPaymentSpendRuleWindowsByRuleIdAfterEach(suite)

					_, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithWindowSeconds(604800).
							Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithWindowSeconds(3600).
							Build())
					assert.NoError(t, err)

					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					if assert.Len(t, windows, 2) {
						assert.Equal(t, 3600, windows[0].WindowSeconds)
						assert.Equal(t, 604800, windows[1].WindowSeconds)
					}
				})
			})

			t.Run("Given the rule is deleted", func(t *testing.T) {
				t.Parallel()

				t.Run("Should cascade-delete its windows", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentSpendRuleWindowsByRuleIdBeforeEach(tc)
					defer WhenListingPaymentSpendRuleWindowsByRuleIdAfterEach(suite)

					_, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(suite.rule.Id).Build())
					assert.NoError(t, err)

					assert.NoError(t, repositories.NewPaymentSpendRuleRepository(suite.dbClient).
						DeleteById(context.Background(), suite.rule.Id))

					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)

					assert.NoError(t, err)
					assert.Empty(t, windows)
				})
			})
		})
	}
}
