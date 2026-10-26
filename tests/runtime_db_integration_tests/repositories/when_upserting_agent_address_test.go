package repositories_test

import (
	"context"
	"testing"

	identity_models "github.com/6022-labs/agentic-agent-node/src/agent_identity/models"
	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type WhenUpsertingAgentAddressTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentAddressRepository
	agentId                  uuid.UUID
}

func WhenUpsertingAgentAddressBeforeEach(tc *common_tests.DatabaseTestCase) *WhenUpsertingAgentAddressTestingSuite {
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

	return &WhenUpsertingAgentAddressTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentAddressRepository(dbClient),
		agentId:                  agentId,
	}
}

func WhenUpsertingAgentAddressAfterEach(suite *WhenUpsertingAgentAddressTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func agentAddress(
	agentId uuid.UUID,
	addressType identity_enums.AddressType,
	address string,
) *identity_models.AgentAddress {
	return &identity_models.AgentAddress{AgentId: agentId, Type: addressType, Address: address}
}

func TestWhenUpsertingAgentAddress(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no address of the type yet", func(t *testing.T) {
				t.Parallel()

				t.Run("Should upsert without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentAddressBeforeEach(tc)
					defer WhenUpsertingAgentAddressAfterEach(suite)

					err := suite.sut.Upsert(
						context.Background(),
						agentAddress(suite.agentId, identity_enums.AddressTypeEvm, "0xabc"),
					)

					assert.NoError(t, err)
				})

				t.Run("Should read it back by agent id and type", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentAddressBeforeEach(tc)
					defer WhenUpsertingAgentAddressAfterEach(suite)
					require.NoError(
						t,
						suite.sut.Upsert(
							context.Background(),
							agentAddress(suite.agentId, identity_enums.AddressTypeEvm, "0xabc"),
						),
					)

					stored, err := suite.sut.GetByAgentIdAndType(
						context.Background(),
						suite.agentId,
						identity_enums.AddressTypeEvm,
					)

					assert.NoError(t, err)
					assert.Equal(t, "0xabc", stored.Address)
				})
			})

			t.Run("Given an address of the type already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should replace it on upsert", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentAddressBeforeEach(tc)
					defer WhenUpsertingAgentAddressAfterEach(suite)

					assert.NoError(
						t,
						suite.sut.Upsert(
							context.Background(),
							agentAddress(suite.agentId, identity_enums.AddressTypeBtc, "bc1qfirst"),
						),
					)
					assert.NoError(
						t,
						suite.sut.Upsert(
							context.Background(),
							agentAddress(suite.agentId, identity_enums.AddressTypeBtc, "bc1qsecond"),
						),
					)

					stored, err := suite.sut.GetByAgentIdAndType(
						context.Background(),
						suite.agentId,
						identity_enums.AddressTypeBtc,
					)

					assert.NoError(t, err)
					assert.Equal(t, "bc1qsecond", stored.Address)
				})
			})
		})
	}
}
