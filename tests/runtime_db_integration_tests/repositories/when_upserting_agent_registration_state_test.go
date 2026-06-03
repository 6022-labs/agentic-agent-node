package repositories_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
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

type WhenUpsertingAgentRegistrationStateTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentRegistrationStateRepository
}

func WhenUpsertingAgentRegistrationStateBeforeEach(tc *repositoryTestCase) *WhenUpsertingAgentRegistrationStateTestingSuite {
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

	return &WhenUpsertingAgentRegistrationStateTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentRegistrationStateRepository(dbClient),
	}
}

func WhenUpsertingAgentRegistrationStateAfterEach(suite *WhenUpsertingAgentRegistrationStateTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingAgentRegistrationState(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no state exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert and return the new state with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					inserted, err := suite.sut.Upsert(model_builders.NewAgentRegistrationStateBuilder().Build())
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)

					var count int64
					suite.dbClient.GetDB().Model(&models.AgentRegistrationState{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given a state already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should advance the mint lifecycle in place and preserve Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					first, err := suite.sut.Upsert(model_builders.NewAgentRegistrationStateBuilder().
						WithMintStatus(identity_enums.TransactionStatusPending).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(model_builders.NewAgentRegistrationStateBuilder().
						WithMintStatus(identity_enums.TransactionStatusConfirmed).
						WithMintChainId(80002).
						WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).
						Build())
					assert.NoError(t, err)
					assert.Equal(t, first.Id, second.Id)
					assert.Equal(t, identity_enums.TransactionStatusConfirmed, second.MintStatus)
					assert.Equal(t, uint64(80002), *second.MintChainId)
					assert.Equal(t, identity_enums.TransactionStatusConfirmed, second.EnsRegistryStatus)

					var count int64
					suite.dbClient.GetDB().Model(&models.AgentRegistrationState{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
