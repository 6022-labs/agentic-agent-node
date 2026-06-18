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
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenGettingPaymentPolicyTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentPolicyRepository
}

func WhenGettingPaymentPolicyBeforeEach(tc *repositoryTestCase) *WhenGettingPaymentPolicyTestingSuite {
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

	return &WhenGettingPaymentPolicyTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentPolicyRepository(dbClient),
	}
}

func WhenGettingPaymentPolicyAfterEach(suite *WhenGettingPaymentPolicyTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingPaymentPolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentPolicyBeforeEach(tc)
					defer WhenGettingPaymentPolicyAfterEach(suite)

					policy, err := suite.sut.Get(context.Background())

					assert.NoError(t, err)
					assert.Nil(t, policy)
				})
			})

			t.Run("Given a policy exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingPaymentPolicyBeforeEach(tc)
					defer WhenGettingPaymentPolicyAfterEach(suite)

					created, err := suite.sut.Create(context.Background(), model_builders.NewPaymentPolicyBuilder().Build())
					assert.NoError(t, err)

					policy, err := suite.sut.Get(context.Background())

					assert.NoError(t, err)
					assert.Equal(t, created.Id, policy.Id)
					assert.Equal(t, created.AgentId, policy.AgentId)
				})
			})
		})
	}
}
