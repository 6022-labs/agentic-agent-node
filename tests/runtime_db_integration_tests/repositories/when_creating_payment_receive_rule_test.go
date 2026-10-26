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

type WhenCreatingPaymentReceiveRuleTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceiveRuleRepository
	policy                   *models.PaymentReceivePolicy
}

func WhenCreatingPaymentReceiveRuleBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenCreatingPaymentReceiveRuleTestingSuite {
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

	policy, err := repositories.NewPaymentReceivePolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}

	return &WhenCreatingPaymentReceiveRuleTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceiveRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenCreatingPaymentReceiveRuleAfterEach(suite *WhenCreatingPaymentReceiveRuleTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given rules exist in insertion order", func(t *testing.T) {
				t.Parallel()

				t.Run("Should generate an id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentReceiveRuleBeforeEach(tc)
					defer WhenCreatingPaymentReceiveRuleAfterEach(suite)

					first, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().
							WithId(uuid.Nil).
							WithPolicyId(suite.policy.Id).
							Build(),
					)

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, first.Id)
				})

				t.Run("Should append the position", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentReceiveRuleBeforeEach(tc)
					defer WhenCreatingPaymentReceiveRuleAfterEach(suite)

					first, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().
							WithId(uuid.Nil).
							WithPolicyId(suite.policy.Id).
							Build(),
					)
					assert.NoError(t, err)

					second, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceiveRuleBuilder().WithPolicyId(suite.policy.Id).Build())

					assert.NoError(t, err)
					assert.Greater(t, second.Position, first.Position)
				})
			})
		})
	}
}
