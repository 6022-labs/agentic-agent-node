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
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenCreatingPaymentSpendPolicyTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendPolicyRepository
}

func WhenCreatingPaymentSpendPolicyBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenCreatingPaymentSpendPolicyTestingSuite {
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

	return &WhenCreatingPaymentSpendPolicyTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendPolicyRepository(dbClient),
	}
}

func WhenCreatingPaymentSpendPolicyAfterEach(suite *WhenCreatingPaymentSpendPolicyTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the head with a generated id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendPolicyBeforeEach(tc)
					defer WhenCreatingPaymentSpendPolicyAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendPolicyBuilder().WithId(uuid.Nil).WithAgentId(agentId).Build())

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, created.Id)

					stored, err := suite.sut.GetByAgentId(context.Background(), created.AgentId)
					assert.NoError(t, err)
					assert.Equal(t, created.Id, stored.Id)
					assert.Equal(t, created.AgentId, stored.AgentId)
				})
			})

			t.Run("Given a policy without payees", func(t *testing.T) {
				t.Parallel()

				// A nil list would serialize as null, which the not-null column refuses.
				t.Run("Should store an empty payee list", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentSpendPolicyBeforeEach(tc)
					defer WhenCreatingPaymentSpendPolicyAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)

					created, err := suite.sut.Create(
						context.Background(),
						model_builders.NewPaymentSpendPolicyBuilder().
							WithAgentId(agentId).
							WithPayToPatterns(nil).
							Build(),
					)
					assert.NoError(t, err)

					stored, err := suite.sut.GetByAgentId(context.Background(), created.AgentId)
					assert.NoError(t, err)
					assert.Equal(t, []string{}, stored.PayToPatterns)
				})
			})
		})
	}
}
