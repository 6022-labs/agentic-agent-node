package configurations

import (
	knowledge_repositories "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"go.uber.org/dig"
	"go.uber.org/zap"
)

func AddAgentKnowledgeDbConfiguration(container *dig.Container) {
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
		logger.Fatal("failed to create agent_knowledge database client", zap.Error(err))
	}
	if err := client.Migrate(); err != nil {
		logger.Fatal("failed to migrate agent_knowledge database", zap.Error(err))
	}
	return client
}
