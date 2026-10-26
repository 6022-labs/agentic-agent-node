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

type WhenCreatingPaymentReceivePolicyTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceivePolicyRepository
}

func WhenCreatingPaymentReceivePolicyBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenCreatingPaymentReceivePolicyTestingSuite {
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

	return &WhenCreatingPaymentReceivePolicyTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceivePolicyRepository(dbClient),
	}
}

func WhenCreatingPaymentReceivePolicyAfterEach(suite *WhenCreatingPaymentReceivePolicyTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentReceivePolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the head with a generated id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentReceivePolicyBeforeEach(tc)
					defer WhenCreatingPaymentReceivePolicyAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentReceivePolicyBuilder().WithId(uuid.Nil).WithAgentId(agentId).Build())

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, created.Id)

					stored, err := suite.sut.GetByAgentId(context.Background(), created.AgentId)
					assert.NoError(t, err)
					assert.Equal(t, created.Id, stored.Id)
					assert.Equal(t, created.AgentId, stored.AgentId)
				})
			})
		})
	}
}
