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

type WhenUpdatingPaymentSpendRuleTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleRepository
	policy                   *models.PaymentSpendPolicy
}

func WhenUpdatingPaymentSpendRuleBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpdatingPaymentSpendRuleTestingSuite {
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

	return &WhenUpdatingPaymentSpendRuleTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenUpdatingPaymentSpendRuleAfterEach(suite *WhenUpdatingPaymentSpendRuleTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace its fields and keep identity and position", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentSpendRuleBeforeEach(tc)
					defer WhenUpdatingPaymentSpendRuleAfterEach(suite)

					rule, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule.Name = "renamed"
					_, err = suite.sut.Update(context.Background(), rule)
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, "renamed", stored.Name)
					assert.Equal(t, rule.Position, stored.Position)
					assert.Equal(t, suite.policy.Id, stored.PolicyId)
				})
			})
		})
	}
}
