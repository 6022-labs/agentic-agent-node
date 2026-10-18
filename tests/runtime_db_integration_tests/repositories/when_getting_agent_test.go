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
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenGettingAgentTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.AgentRepository
}

func WhenGettingAgentBeforeEach(tc *repositoryTestCase) *WhenGettingAgentTestingSuite {
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

	return &WhenGettingAgentTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewAgentRepository(dbClient),
	}
}

func WhenGettingAgentAfterEach(suite *WhenGettingAgentTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenGettingAgent(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no agent exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return nil without error", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentBeforeEach(tc)
					defer WhenGettingAgentAfterEach(suite)

					got, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Nil(t, got)
				})
			})

			t.Run("Given an agent exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return the persisted row", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentBeforeEach(tc)
					defer WhenGettingAgentAfterEach(suite)

					inserted, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert", "trait": "curious"}).
						WithEnsDomain("ada.agent.eth").
						WithImages(map[string]string{"avatar": "https://example.com/a.png"}).
						Build())

					assert.NoError(t, err)

					got, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, inserted.Id, got.Id)
					assert.Equal(t, "Ada", lo.FromPtr(got.Name))
					assert.Equal(t, "ada.agent.eth", lo.FromPtr(got.EnsDomain))
				})

				t.Run("Should return its attributes JSON map", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentBeforeEach(tc)
					defer WhenGettingAgentAfterEach(suite)

					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert", "trait": "curious"}).
						WithEnsDomain("ada.agent.eth").
						WithImages(map[string]string{"avatar": "https://example.com/a.png"}).
						Build())

					assert.NoError(t, err)

					got, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, "expert", got.Attributes["role"])
					assert.Equal(t, "curious", got.Attributes["trait"])
				})

				t.Run("Should return its images JSON map", func(t *testing.T) {
					t.Parallel()
					suite := WhenGettingAgentBeforeEach(tc)
					defer WhenGettingAgentAfterEach(suite)

					_, err := suite.sut.Upsert(context.Background(), model_builders.NewAgentBuilder().
						WithName("Ada").
						WithAttributes(map[string]string{"role": "expert", "trait": "curious"}).
						WithEnsDomain("ada.agent.eth").
						WithImages(map[string]string{"avatar": "https://example.com/a.png"}).
						Build())

					assert.NoError(t, err)

					got, err := suite.sut.Get(context.Background())
					assert.NoError(t, err)
					assert.Equal(t, "https://example.com/a.png", got.Images["avatar"])
				})
			})

		})
	}
}
