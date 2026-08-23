package configurations

import (
	identity_repositories "github.com/6022-labs/agentic-agent-node/src/agent_identity/repositories"
	knowledge_enums "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	knowledge_repositories "github.com/6022-labs/agentic-agent-node/src/agent_knowledge/repositories"
	payment_repositories "github.com/6022-labs/agentic-agent-node/src/agent_payment/repositories"
	reasoning_repositories "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/metrics"
	metrics_interfaces "github.com/6022-labs/agentic-agent-node/src/runtime_db/metrics/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
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
	if err := container.Provide(
		func(databaseSettings *settings.DatabaseSettings) *metrics.DbQueryMetrics {
			return metrics.NewDbQueryMetrics(dbSystemName(databaseSettings.Driver))
		},
		dig.As(new(metrics_interfaces.DbQueryMetricsInterface)),
	); err != nil {
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
		repositories.NewConversationRepository,
		dig.As(new(knowledge_repositories.ConversationRepositoryInterface)),
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
		repositories.NewSelfToolRepository,
		dig.As(new(reasoning_repositories.SelfToolRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewAgentGuidanceRepository,
		dig.As(new(reasoning_repositories.AgentGuidanceRepositoryInterface)),
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
	if err := container.Provide(
		repositories.NewAgentAddressRepository,
		dig.As(new(identity_repositories.AgentAddressRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewPaymentPolicyRepository,
		dig.As(new(payment_repositories.PaymentPolicyRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewPaymentRuleRepository,
		dig.As(new(payment_repositories.PaymentRuleRepositoryInterface)),
	); err != nil {
		panic(err)
	}
	if err := container.Provide(
		repositories.NewPaymentRuleX402ConfigRepository,
		dig.As(new(payment_repositories.PaymentRuleX402ConfigRepositoryInterface)),
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
	dbMetrics metrics_interfaces.DbQueryMetricsInterface,
) *runtime_db.DatabaseClient {
	client, err := runtime_db.NewDatabaseClient(dbSettings, migrationsSettings, logger)
	if err != nil {
		logger.Fatal("failed to create runtime database client", zap.Error(err))
	}
	if err := client.GetDB().Use(metrics.NewGormMetricsPlugin(dbMetrics)); err != nil {
		logger.Fatal("failed to register db metrics plugin", zap.Error(err))
	}
	if err := client.Migrate(); err != nil {
		logger.Fatal("failed to migrate runtime database", zap.Error(err))
	}
	return client
}

// dbSystemName maps the configured driver to its semconv db.system.name value.
func dbSystemName(driver knowledge_enums.DatabaseDriver) attribute.KeyValue {
	switch driver {
	case knowledge_enums.DatabaseDriverMySQL:
		return semconv.DBSystemNameMySQL
	case knowledge_enums.DatabaseDriverMariaDB:
		return semconv.DBSystemNameMariaDB
	case knowledge_enums.DatabaseDriverSqlite:
		return semconv.DBSystemNameSQLite
	default:
		return semconv.DBSystemNamePostgreSQL
	}
}
