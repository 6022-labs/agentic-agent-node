package repositories_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"

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

type WhenGettingPaymentSpendPolicyByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendPolicyRepository
}

func WhenGettingPaymentSpendPolicyByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingPaymentSpendPolicyByAgentIdTestingSuite {
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

	return &WhenGettingPaymentSpendPolicyByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendPolicyRepository(dbClient),
	}
}

func WhenGettingPaymentSpendPolicyByAgentIdAfterEach(suite *WhenGettingPaymentSpendPolicyByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentSpendPolicyByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendPolicyByAgentIdBeforeEach(tc)
					defer WhenGettingPaymentSpendPolicyByAgentIdAfterEach(suite)

					policy, err := suite.sut.GetByAgentId(context.Background(), uuid.New())

					assert.NoError(t, err)
					assert.Nil(t, policy)
				})
			})

			t.Run("Given a policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentSpendPolicyByAgentIdBeforeEach(tc)
					defer WhenGettingPaymentSpendPolicyByAgentIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					created, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)

					policy, err := suite.sut.GetByAgentId(context.Background(), created.AgentId)

					assert.NoError(t, err)
					assert.Equal(t, created.Id, policy.Id)
					assert.Equal(t, created.AgentId, policy.AgentId)
				})
			})
		})
	}
}
