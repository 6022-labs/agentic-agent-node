package repositories_test

import (
	"context"
	"testing"

	reasoning_enums "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models/enums"
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

type WhenIsolatingByAgentTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	llmConnectionRepository  *repositories.LLMConnectionRepository
	mcpConnectionRepository  *repositories.MCPConnectionRepository
}

func WhenIsolatingByAgentBeforeEach(tc *repositoryTestCase) *WhenIsolatingByAgentTestingSuite {
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

	return &WhenIsolatingByAgentTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		llmConnectionRepository:  repositories.NewLLMConnectionRepository(dbClient),
		mcpConnectionRepository:  repositories.NewMCPConnectionRepository(dbClient),
	}
}

func WhenIsolatingByAgentAfterEach(suite *WhenIsolatingByAgentTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenIsolatingByAgent(t *testing.T) {
	t.Parallel()

	for _, tc := range repositoryTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given two agents each own an llm and an mcp connection", func(t *testing.T) {
				t.Parallel()

				t.Run("Should only return the rows scoped to the queried agent", func(t *testing.T) {
					t.Parallel()
					suite := WhenIsolatingByAgentBeforeEach(tc)
					defer WhenIsolatingByAgentAfterEach(suite)

					agentA := uuid.New()
					agentB := uuid.New()

					llmA, err := suite.llmConnectionRepository.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentA).
						WithPlatform(reasoning_enums.LLMConnectionPlatformOpenAI).
						WithModel("gpt-4o").
						Build())
					assert.NoError(t, err)

					llmB, err := suite.llmConnectionRepository.Upsert(context.Background(), model_builders.NewLLMConnectionBuilder().
						WithAgentId(agentB).
						WithPlatform(reasoning_enums.LLMConnectionPlatformAnthropic).
						WithModel("claude").
						Build())
					assert.NoError(t, err)

					mcpA := model_builders.NewMCPConnectionBuilder().WithAgentId(agentA).WithName("a-tool").Build()
					_, err = suite.mcpConnectionRepository.Create(context.Background(), mcpA)
					assert.NoError(t, err)
					mcpB := model_builders.NewMCPConnectionBuilder().WithAgentId(agentB).WithName("b-tool").Build()
					_, err = suite.mcpConnectionRepository.Create(context.Background(), mcpB)
					assert.NoError(t, err)

					gotA, err := suite.llmConnectionRepository.GetByAgentId(context.Background(), agentA)
					assert.NoError(t, err)
					assert.NotNil(t, gotA)
					assert.Equal(t, llmA.Id, gotA.Id)
					assert.Equal(t, agentA, gotA.AgentId)

					gotB, err := suite.llmConnectionRepository.GetByAgentId(context.Background(), agentB)
					assert.NoError(t, err)
					assert.NotNil(t, gotB)
					assert.Equal(t, llmB.Id, gotB.Id)
					assert.Equal(t, agentB, gotB.AgentId)
					assert.NotEqual(t, gotA.Id, gotB.Id)

					listA, err := suite.mcpConnectionRepository.ListByAgentId(context.Background(), agentA)
					assert.NoError(t, err)
					assert.Len(t, listA, 1)
					assert.Equal(t, "a-tool", listA[0].Name)
					assert.Equal(t, agentA, listA[0].AgentId)

					listB, err := suite.mcpConnectionRepository.ListByAgentId(context.Background(), agentB)
					assert.NoError(t, err)
					assert.Len(t, listB, 1)
					assert.Equal(t, "b-tool", listB[0].Name)
					assert.Equal(t, agentB, listB[0].AgentId)
				})
			})
		})
	}
}
