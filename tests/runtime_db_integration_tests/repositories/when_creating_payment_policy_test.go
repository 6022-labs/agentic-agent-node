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

type WhenCreatingPaymentPolicyTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentPolicyRepository
}

func WhenCreatingPaymentPolicyBeforeEach(tc *repositoryTestCase) *WhenCreatingPaymentPolicyTestingSuite {
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

	return &WhenCreatingPaymentPolicyTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentPolicyRepository(dbClient),
	}
}

func WhenCreatingPaymentPolicyAfterEach(suite *WhenCreatingPaymentPolicyTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenCreatingPaymentPolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should persist the head with a generated id", func(t *testing.T) {
					t.Parallel()
					suite := WhenCreatingPaymentPolicyBeforeEach(tc)
					defer WhenCreatingPaymentPolicyAfterEach(suite)

					created, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentPolicyBuilder().WithId(uuid.Nil).Build())

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
