package repositories_test

import (
	"context"
	"strings"
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

type WhenGettingConversationByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenGettingConversationByIdBeforeEach(tc *common_tests.DatabaseTestCase) *WhenGettingConversationByIdTestingSuite {
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

	return &WhenGettingConversationByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenGettingConversationByIdAfterEach(suite *WhenGettingConversationByIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingConversationById(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a conversation exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return it by its id", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdBeforeEach(tc)
					defer WhenGettingConversationByIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Ensure(context.Background(),
						model_builders.NewConversationBuilder().WithId("conv-mine").WithAgentId(agentId).Build())
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), "conv-mine")

					assert.NoError(t, err)
					if assert.NotNil(t, got) {
						assert.Equal(t, "conv-mine", got.Id)
					}
				})

				t.Run("Should carry its owner agent id", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdBeforeEach(tc)
					defer WhenGettingConversationByIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					_, err := suite.sut.Ensure(context.Background(),
						model_builders.NewConversationBuilder().WithId("conv-mine").WithAgentId(agentId).Build())
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), "conv-mine")

					assert.NoError(t, err)
					if assert.NotNil(t, got) {
						assert.Equal(t, agentId, got.AgentId)
					}
				})

				t.Run("Should carry its caller id", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdBeforeEach(tc)
					defer WhenGettingConversationByIdAfterEach(suite)

					agentId := uuid.New()
					seeds.SeedAgent(suite.dbClient, agentId)
					alice := strings.ToLower(common_tests.RandomEthAddress().Hex())
					_, err := suite.sut.Ensure(context.Background(), model_builders.NewConversationBuilder().
						WithId("conv-mine").
						WithAgentId(agentId).
						WithCallerId(alice).
						Build())
					assert.NoError(t, err)

					got, err := suite.sut.GetById(context.Background(), "conv-mine")

					assert.NoError(t, err)
					if assert.NotNil(t, got) && assert.NotNil(t, got.CallerId) {
						assert.Equal(t, alice, *got.CallerId)
					}
				})
			})

			t.Run("Given the conversation does not exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdBeforeEach(tc)
					defer WhenGettingConversationByIdAfterEach(suite)

					got, err := suite.sut.GetById(context.Background(), "missing")

					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})
		})
	}
}
