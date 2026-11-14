package repositories_test

import (
	"context"
	"testing"

	identity_enums "github.com/6022-labs/agentic-agent-node/src/agent_identity/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenGettingAllAgentAddressesByAgentIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentAddressRepository
	agentId                  uuid.UUID
}

func WhenGettingAllAgentAddressesByAgentIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenGettingAllAgentAddressesByAgentIdTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	return &WhenGettingAllAgentAddressesByAgentIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentAddressRepository(dbClient),
		agentId:                  agentId,
	}
}

func WhenGettingAllAgentAddressesByAgentIdAfterEach(suite *WhenGettingAllAgentAddressesByAgentIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingAllAgentAddressesByAgentId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given stored addresses", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them keyed by type", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAllAgentAddressesByAgentIdBeforeEach(tc)
					defer WhenGettingAllAgentAddressesByAgentIdAfterEach(suite)

					assert.NoError(
						t,
						suite.sut.Upsert(
							context.Background(),
							agentAddress(suite.agentId, identity_enums.AddressTypeEvm, "0xabc"),
						),
					)
					assert.NoError(
						t,
						suite.sut.Upsert(
							context.Background(),
							agentAddress(suite.agentId, identity_enums.AddressTypeBtc, "bc1qxyz"),
						),
					)

					addresses, err := suite.sut.GetAllByAgentId(context.Background(), suite.agentId)

					assert.NoError(t, err)
					assert.Equal(t, map[identity_enums.AddressType]string{
						identity_enums.AddressTypeEvm: "0xabc",
						identity_enums.AddressTypeBtc: "bc1qxyz",
					}, addresses)
				})
			})

			t.Run("Given no stored addresses", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty map", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAllAgentAddressesByAgentIdBeforeEach(tc)
					defer WhenGettingAllAgentAddressesByAgentIdAfterEach(suite)

					addresses, err := suite.sut.GetAllByAgentId(context.Background(), suite.agentId)

					assert.NoError(t, err)
					assert.Empty(t, addresses)
					assert.NotNil(t, addresses)
				})
			})
		})
	}
}
