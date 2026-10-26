package repositories_test

import (
	"context"
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

func WhenUpsertingAgentRegistrationStateBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpsertingAgentRegistrationStateTestingSuite {
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

	return &WhenUpsertingAgentRegistrationStateTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentRegistrationStateRepository(dbClient),
	}
}

func WhenUpsertingAgentRegistrationStateAfterEach(suite *WhenUpsertingAgentRegistrationStateTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingAgentRegistrationState(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no state exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the new state with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					inserted, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().WithAgentId(agentId).Build(),
					)

					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)
				})

				t.Run("Should store a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().WithAgentId(agentId).Build(),
					)
					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.AgentRegistrationState{}).Count(&count)

					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given a state already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should preserve its Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					agentId := uuid.New()
					if err := suite.dbClient.GetDB(context.Background()).
						Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
						panic(err)
					}

					first, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusPending).
							Build(),
					)
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusConfirmed).
							WithMintChainId(80002).
							WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).
							Build(),
					)

					assert.NoError(t, err)
					assert.Equal(t, first.Id, second.Id)
				})

				t.Run("Should advance the mint status", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusPending).
							Build(),
					)
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusConfirmed).
							WithMintChainId(80002).
							WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).
							Build(),
					)

					assert.NoError(t, err)
					assert.Equal(t, identity_enums.TransactionStatusConfirmed, second.MintStatus)
				})

				t.Run("Should record the mint chain id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusPending).
							Build(),
					)
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusConfirmed).
							WithMintChainId(80002).
							WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).
							Build(),
					)

					assert.NoError(t, err)
					assert.Equal(t, uint64(80002), *second.MintChainId)
				})

				t.Run("Should advance the ENS registry status", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusPending).
							Build(),
					)
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusConfirmed).
							WithMintChainId(80002).
							WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).
							Build(),
					)

					assert.NoError(t, err)
					assert.Equal(t, identity_enums.TransactionStatusConfirmed, second.EnsRegistryStatus)
				})

				t.Run("Should keep a single row", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentRegistrationStateBeforeEach(tc)
					defer WhenUpsertingAgentRegistrationStateAfterEach(suite)

					agentId := uuid.New()
					seedAgent(suite.dbClient, agentId)

					_, err := suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusPending).
							Build(),
					)
					assert.NoError(t, err)

					_, err = suite.sut.Upsert(
						context.Background(),
						model_builders.NewAgentRegistrationStateBuilder().
							WithAgentId(agentId).
							WithMintStatus(identity_enums.TransactionStatusConfirmed).
							WithMintChainId(80002).
							WithEnsRegistryStatus(identity_enums.TransactionStatusConfirmed).
							Build(),
					)

					assert.NoError(t, err)

					var count int64
					suite.dbClient.GetDB(context.Background()).Model(&models.AgentRegistrationState{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
