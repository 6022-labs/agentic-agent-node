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

type WhenUpsertingAgentTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentRepository
}

func WhenUpsertingAgentBeforeEach(tc *repositoryTestCase) *WhenUpsertingAgentTestingSuite {
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

	return &WhenUpsertingAgentTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentRepository(dbClient),
	}
}

func WhenUpsertingAgentAfterEach(suite *WhenUpsertingAgentTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpsertingAgent(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no agent exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should insert and return the new agent with a generated Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					inserted, err := suite.sut.Upsert(model_builders.NewAgentBuilder().
						WithName("Ada").
						WithRole(identity_enums.AgentRoleExpert).
						Build())
					assert.NoError(t, err)
					assert.NotEqual(t, uuid.Nil, inserted.Id)

					var count int64
					suite.dbClient.GetDB().Model(&models.Agent{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})

			t.Run("Given an agent already exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should update the existing row in place and preserve Id", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpsertingAgentBeforeEach(tc)
					defer WhenUpsertingAgentAfterEach(suite)

					first, err := suite.sut.Upsert(model_builders.NewAgentBuilder().
						WithName("Ada").
						WithRole(identity_enums.AgentRoleExpert).
						Build())
					assert.NoError(t, err)

					second, err := suite.sut.Upsert(model_builders.NewAgentBuilder().
						WithName("Babbage").
						WithRole(identity_enums.AgentRoleFacilitator).
						Build())
					assert.NoError(t, err)
					assert.Equal(t, first.Id, second.Id)
					assert.Equal(t, "Babbage", second.Name)
					assert.Equal(t, identity_enums.AgentRoleFacilitator, second.Role)

					var count int64
					suite.dbClient.GetDB().Model(&models.Agent{}).Count(&count)
					assert.Equal(t, int64(1), count)
				})
			})
		})
	}
}
