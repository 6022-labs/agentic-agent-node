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

type WhenDeletingPaymentRuleByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentRuleRepository
	policy                   *models.PaymentPolicy
}

func WhenDeletingPaymentRuleByIdBeforeEach(tc *repositoryTestCase) *WhenDeletingPaymentRuleByIdTestingSuite {
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

	policy, err := repositories.NewPaymentPolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentPolicyBuilder().Build())
	if err != nil {
		panic(err)
	}

	return &WhenDeletingPaymentRuleByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenDeletingPaymentRuleByIdAfterEach(suite *WhenDeletingPaymentRuleByIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentRuleById(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentRuleByIdBeforeEach(tc)
					defer WhenDeletingPaymentRuleByIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteById(context.Background(), uuid.New()))
				})
			})

			t.Run("Given several rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove only the targeted rule", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentRuleByIdBeforeEach(tc)
					defer WhenDeletingPaymentRuleByIdAfterEach(suite)

					first, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)
					second, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(suite.policy.Id).Build())
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
