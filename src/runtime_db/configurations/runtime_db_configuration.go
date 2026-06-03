package configurations

import (
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	knowledge_repositories "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	reasoning_repositories "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddRuntimeDbConfiguration(container *dig.Container) {
	if err := container.Provide(settings.NewDatabaseSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(settings.NewMigrationsSettings); err != nil {
		panic(err)
	}
	if err := container.Provide(newDatabaseClient); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewConversationMessageRepository,
		dig.As(new(knowledge_repositories.ConversationMessageRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewLLMConnectionRepository,
		dig.As(new(reasoning_repositories.LLMConnectionRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewMCPConnectionRepository,
		dig.As(new(reasoning_repositories.MCPConnectionRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewAgentRepository,
		dig.As(new(identity_repositories.AgentRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewAgentRegistrationStateRepository,
		dig.As(new(identity_repositories.AgentRegistrationStateRepositoryInterface)),
	); err != nil {
		panic(err)
	}
}

// newDatabaseClient opens the DB and runs migrations at boot. A failure here
// is fatal — gated routes can't serve without the local store.
func newDatabaseClient(
	dbSettings *settings.DatabaseSettings,
	migrationsSettings *settings.MigrationsSettings,
	logger *zap.Logger,
) *runtime_db.DatabaseClient {
	client, err := runtime_db.NewDatabaseClient(dbSettings, migrationsSettings, logger)
	if err != nil {
		logger.Fatal("failed to create runtime database client", zap.Error(err))
	}
	if err := client.Migrate(); err != nil {
		logger.Fatal("failed to migrate runtime database", zap.Error(err))
	}
	return client
}
