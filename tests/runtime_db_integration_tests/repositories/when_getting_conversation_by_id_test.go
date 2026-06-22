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

type WhenGettingConversationByIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.ConversationRepository
}

func WhenGettingConversationByIdBeforeEach(tc *repositoryTestCase) *WhenGettingConversationByIdTestingSuite {
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

	return &WhenGettingConversationByIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewConversationRepository(dbClient),
	}
}

func WhenGettingConversationByIdAfterEach(suite *WhenGettingConversationByIdTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingConversationById(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a conversation exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should be returned by its id, carrying its owner AgentId", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingConversationByIdBeforeEach(tc)
					defer WhenGettingConversationByIdAfterEach(suite)

					ownerAgentId := uuid.New()
					seedAgent(suite.dbClient, ownerAgentId)
					_, ensureErr := suite.sut.Ensure(context.Background(),
						model_builders.NewConversationBuilder().WithId("conv-mine").WithAgentId(ownerAgentId).Build())
					assert.NoError(t, ensureErr)

					got, err := suite.sut.GetById(context.Background(), "conv-mine")
					assert.NoError(t, err)
					assert.NotNil(t, got)
					assert.Equal(t, "conv-mine", got.Id)
					assert.Equal(t, ownerAgentId, got.AgentId)
				})

				t.Run("Should return nil for an unknown id", func(t *testing.T) {
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
