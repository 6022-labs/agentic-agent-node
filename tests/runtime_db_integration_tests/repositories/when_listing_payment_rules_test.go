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

type WhenListingPaymentRulesTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentRuleRepository
	policy                   *models.PaymentPolicy
}

func WhenListingPaymentRulesBeforeEach(tc *repositoryTestCase) *WhenListingPaymentRulesTestingSuite {
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

	return &WhenListingPaymentRulesTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenListingPaymentRulesAfterEach(suite *WhenListingPaymentRulesTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingPaymentRules(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty list", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentRulesBeforeEach(tc)
					defer WhenListingPaymentRulesAfterEach(suite)

					rules, err := suite.sut.List(context.Background())

					assert.NoError(t, err)
					assert.Empty(t, rules)
				})
			})

			t.Run("Given several rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them in creation order", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentRulesBeforeEach(tc)
					defer WhenListingPaymentRulesAfterEach(suite)

					_, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(suite.policy.Id).WithName("first").Build())
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(suite.policy.Id).WithName("second").Build())
					assert.NoError(t, err)

					rules, err := suite.sut.List(context.Background())

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
