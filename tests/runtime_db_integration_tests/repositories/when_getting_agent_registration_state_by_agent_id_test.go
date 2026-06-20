package repositories_test

import (
	"context"
	"testing"

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

type WhenGettingAgentRegistrationStateByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentRegistrationStateRepository
}

func WhenGettingAgentRegistrationStateByAgentIdBeforeEach(tc *repositoryTestCase) *WhenGettingAgentRegistrationStateByAgentIdTestingSuite {
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

	return &WhenGettingAgentRegistrationStateByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentRegistrationStateRepository(dbClient),
	}
}

func WhenGettingAgentRegistrationStateByAgentIdAfterEach(suite *WhenGettingAgentRegistrationStateByAgentIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingAgentRegistrationStateByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no state has been persisted", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentRegistrationStateByAgentIdBeforeEach(tc)
					defer WhenGettingAgentRegistrationStateByAgentIdAfterEach(suite)

					got, err := suite.sut.GetByAgentId(context.Background(), uuid.New())
					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})

			t.Run("Given a state has been persisted with mint detail", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the persisted row including nullable chain fields", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentRegistrationStateByAgentIdBeforeEach(tc)
					defer WhenGettingAgentRegistrationStateByAgentIdAfterEach(suite)

					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentRegistrationStateBuilder().
						WithMintStatus(identity_enums.TransactionStatusConfirmed).
						WithMintChainId(80002).
						WithMintTxHash("0xabc").
						WithEnsRegistryStatus(identity_enums.TransactionStatusPending).
						Build())

					assert.NoError(t, err)

					got, err := suite.sut.GetByAgentId(context.Background(), inserted.AgentId)
					assert.NoError(t, err)
					assert.Equal(t, inserted.Id, got.Id)
					assert.Equal(t, identity_enums.TransactionStatusConfirmed, got.MintStatus)
					assert.Equal(t, uint64(80002), *got.MintChainId)
					assert.Equal(t, "0xabc", *got.MintTxHash)
					assert.Equal(t, identity_enums.TransactionStatusPending, got.EnsRegistryStatus)
					assert.Nil(t, got.EnsTxHash)
				})
			})
		})
	}
}
