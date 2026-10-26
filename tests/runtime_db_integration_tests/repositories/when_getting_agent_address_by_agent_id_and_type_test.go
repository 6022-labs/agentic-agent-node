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
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenGettingAgentAddressByAgentIdAndTypeTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentAddressRepository
	agentId                  uuid.UUID
}

func WhenGettingAgentAddressByAgentIdAndTypeBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingAgentAddressByAgentIdAndTypeTestingSuite {
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

	return &WhenGettingAgentAddressByAgentIdAndTypeTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentAddressRepository(dbClient),
		agentId:                  agentId,
	}
}

func WhenGettingAgentAddressByAgentIdAndTypeAfterEach(suite *WhenGettingAgentAddressByAgentIdAndTypeTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingAgentAddressByAgentIdAndType(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a stored address of the type", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it for the matching agent and type", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentAddressByAgentIdAndTypeBeforeEach(tc)
					defer WhenGettingAgentAddressByAgentIdAndTypeAfterEach(suite)

					assert.NoError(
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

			t.Run("Given no address of the requested type", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentAddressByAgentIdAndTypeBeforeEach(tc)
					defer WhenGettingAgentAddressByAgentIdAndTypeAfterEach(suite)

					stored, err := suite.sut.GetByAgentIdAndType(
						context.Background(),
						suite.agentId,
						identity_enums.AddressTypeBtc,
					)

					assert.NoError(t, err)
					assert.Nil(t, stored)
				})
			})
		})
	}
}
