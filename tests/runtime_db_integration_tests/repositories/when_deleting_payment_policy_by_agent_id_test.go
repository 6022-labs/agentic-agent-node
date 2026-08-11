package repositories_test

import (
	"context"
	"testing"

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

type WhenDeletingPaymentPolicyByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentPolicyRepository
	paymentRuleRepository    *repositories.PaymentRuleRepository
}

func WhenDeletingPaymentPolicyByAgentIdBeforeEach(
	tc *repositoryTestCase,
) *WhenDeletingPaymentPolicyByAgentIdTestingSuite {
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

	return &WhenDeletingPaymentPolicyByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentPolicyRepository(dbClient),
		paymentRuleRepository:    repositories.NewPaymentRuleRepository(dbClient),
	}
}

func WhenDeletingPaymentPolicyByAgentIdAfterEach(suite *WhenDeletingPaymentPolicyByAgentIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenDeletingPaymentPolicyByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be a no-op", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentPolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentPolicyByAgentIdAfterEach(suite)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), uuid.New()))
				})
			})

			t.Run("Given a policy with rules exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should remove the head and its rules", func(t *testing.T) {
					t.Parallel()
					suite := WhenDeletingPaymentPolicyByAgentIdBeforeEach(tc)
					defer WhenDeletingPaymentPolicyByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					policy, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentPolicyBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)
					_, err = suite.paymentRuleRepository.Create(
						context.Background(),
						model_builders.NewPaymentRuleBuilder().WithPolicyId(policy.Id).Build(),
					)
					assert.NoError(t, err)

					assert.NoError(t, suite.sut.DeleteByAgentId(context.Background(), policy.AgentId))

					gone, err := suite.sut.GetByAgentId(context.Background(), policy.AgentId)
					assert.NoError(t, err)
					assert.Nil(t, gone)

					rules, err := suite.paymentRuleRepository.ListByPolicyId(context.Background(), policy.Id)
					assert.NoError(t, err)
					assert.Empty(t, rules)
				})
			})
		})
	}
}
