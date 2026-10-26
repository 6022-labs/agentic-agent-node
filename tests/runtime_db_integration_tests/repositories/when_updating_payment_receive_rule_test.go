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

type WhenUpdatingPaymentReceiveRuleTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceiveRuleRepository
	policy                   *models.PaymentReceivePolicy
}

func WhenUpdatingPaymentReceiveRuleBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpdatingPaymentReceiveRuleTestingSuite {
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

	return &WhenUpdatingPaymentReceiveRuleTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceiveRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenUpdatingPaymentReceiveRuleAfterEach(suite *WhenUpdatingPaymentReceiveRuleTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace its fields", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentReceiveRuleBeforeEach(tc)
					defer WhenUpdatingPaymentReceiveRuleAfterEach(suite)

					rule, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule.Name = "renamed"
					_, err = suite.sut.Update(context.Background(), rule)
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, "renamed", stored.Name)
				})

				t.Run("Should keep its position", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentReceiveRuleBeforeEach(tc)
					defer WhenUpdatingPaymentReceiveRuleAfterEach(suite)

					rule, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule.Name = "renamed"
					_, err = suite.sut.Update(context.Background(), rule)
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, rule.Position, stored.Position)
				})

				t.Run("Should keep its policy", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentReceiveRuleBeforeEach(tc)
					defer WhenUpdatingPaymentReceiveRuleAfterEach(suite)

					rule, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule.Name = "renamed"
					_, err = suite.sut.Update(context.Background(), rule)
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, suite.policy.Id, stored.PolicyId)
				})
			})
		})
	}
}
