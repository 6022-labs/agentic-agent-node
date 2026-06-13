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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenUpdatingPaymentRuleTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentRuleRepository
	policy                   *models.PaymentPolicy
}

func WhenUpdatingPaymentRuleBeforeEach(tc *repositoryTestCase) *WhenUpdatingPaymentRuleTestingSuite {
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

	return &WhenUpdatingPaymentRuleTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenUpdatingPaymentRuleAfterEach(suite *WhenUpdatingPaymentRuleTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingPaymentRule(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the rule exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace its fields and keep identity and position", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentRuleBeforeEach(tc)
					defer WhenUpdatingPaymentRuleAfterEach(suite)

					rule, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(suite.policy.Id).Build())
					assert.NoError(t, err)

					rule.Name = "renamed"
					rule.AddressPatterns = []string{".*"}
					_, err = suite.sut.Update(context.Background(), rule)
					assert.NoError(t, err)

					stored, err := suite.sut.GetById(context.Background(), rule.Id)
					assert.NoError(t, err)
					assert.Equal(t, "renamed", stored.Name)
					assert.Equal(t, []string{".*"}, stored.AddressPatterns)
					assert.Equal(t, rule.Position, stored.Position)
					assert.Equal(t, suite.policy.Id, stored.PolicyId)
				})
			})
		})
	}
}
