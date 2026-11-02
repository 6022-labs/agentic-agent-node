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

type WhenCreatingPaymentSpendRuleWindowTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleWindowRepository
	rule                     *models.PaymentSpendRule
}

func WhenCreatingPaymentSpendRuleWindowBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenCreatingPaymentSpendRuleWindowTestingSuite {
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

	return &WhenCreatingPaymentSpendRuleWindowTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleWindowRepository(dbClient),
		rule:                     rule,
	}
}

func WhenCreatingPaymentSpendRuleWindowAfterEach(suite *WhenCreatingPaymentSpendRuleWindowTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentSpendRuleWindow(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a window for a rule", func(t *testing.T) {
				t.Parallel()

				t.Run("Should generate an id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendRuleWindowBeforeEach(tc)
					defer WhenCreatingPaymentSpendRuleWindowAfterEach(suite)

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithId(uuid.Nil).
							WithPaymentSpendRuleId(suite.rule.Id).
							Build())
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, created.Id)
				})

				t.Run("Should start with nothing spent", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendRuleWindowBeforeEach(tc)
					defer WhenCreatingPaymentSpendRuleWindowAfterEach(suite)

					_, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							Build())
					assert.NoError(t, err)

					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					if assert.Len(t, windows, 1) {
						assert.Equal(t, "0", windows[0].Spent)
						assert.Nil(t, windows[0].StartedAt)
					}
				})
			})

			t.Run("Given the rule already has a window of that length", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse the second window", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendRuleWindowBeforeEach(tc)
					defer WhenCreatingPaymentSpendRuleWindowAfterEach(suite)

					_, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithWindowSeconds(86400).
							Build())
					assert.NoError(t, err)

					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithWindowSeconds(86400).
							Build())

					// Drivers word the unique violation differently, so the proof is that nothing was stored.
					assert.Error(t, err)
					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					assert.Len(t, windows, 1)
				})
			})
		})
	}
}
