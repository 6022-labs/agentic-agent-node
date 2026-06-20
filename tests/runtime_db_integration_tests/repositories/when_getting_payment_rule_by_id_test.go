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

type WhenGettingPaymentRuleByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentRuleRepository
	policy                   *models.PaymentPolicy
}

func WhenGettingPaymentRuleByIdBeforeEach(tc *repositoryTestCase) *WhenGettingPaymentRuleByIdTestingSuite {
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

	return &WhenGettingPaymentRuleByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenGettingPaymentRuleByIdAfterEach(suite *WhenGettingPaymentRuleByIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentRuleById(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentRuleByIdBeforeEach(tc)
					defer WhenGettingPaymentRuleByIdAfterEach(suite)

					rule, err := suite.sut.GetById(context.Background(), uuid.New())

					assert.NoError(t, err)
					assert.Nil(t, rule)
				})
			})

			t.Run("Given the rule exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it with the patterns decoded", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentRuleByIdBeforeEach(tc)
					defer WhenGettingPaymentRuleByIdAfterEach(suite)

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule, err := suite.sut.GetById(context.Background(), created.Id)

					assert.NoError(t, err)
					assert.Equal(t, created.Name, rule.Name)
					assert.Equal(t, created.Method, rule.Method)
				})
			})
		})
	}
}
