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

type WhenCreatingPaymentRuleTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentRuleRepository
	policy                   *models.PaymentPolicy
}

func WhenCreatingPaymentRuleBeforeEach(tc *repositoryTestCase) *WhenCreatingPaymentRuleTestingSuite {
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

	return &WhenCreatingPaymentRuleTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenCreatingPaymentRuleAfterEach(suite *WhenCreatingPaymentRuleTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentRule(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given rules are created one after another", func(t *testing.T) {
				t.Parallel()

				t.Run("Should generate ids and append positions", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentRuleBeforeEach(tc)
					defer WhenCreatingPaymentRuleAfterEach(suite)

					first, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithId(uuid.Nil).WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, first.Id)

					second, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)
					assert.Greater(t, second.Position, first.Position)
				})
			})
		})
	}
}
