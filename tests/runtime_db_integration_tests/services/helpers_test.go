package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	runtime_db_services "github.com/6022-labs/agentic-agent-node/src/runtime_db/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type unitOfWorkTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *runtime_db_services.UnitOfWorkFactory
	connections              *repositories.MCPConnectionRepository
	agentId                  uuid.UUID
}

func newUnitOfWorkTestingSuite(tc *common_tests.DatabaseTestCase) *unitOfWorkTestingSuite {
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
	if err := dbClient.GetDB(context.Background()).
		Create(model_builders.NewAgentBuilder().WithId(agentId).Build()).Error; err != nil {
		panic(err)
	}

	return &unitOfWorkTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      runtime_db_services.NewUnitOfWorkFactory(dbClient),
		connections:              repositories.NewMCPConnectionRepository(dbClient),
		agentId:                  agentId,
	}
}

func (s *unitOfWorkTestingSuite) close() {
	_ = s.dbClient.Close()
	s.databaseContainerBuilder.DropDatabase(s.dbName)
}

func (s *unitOfWorkTestingSuite) seedConnection(t *testing.T) *models.MCPConnection {
	t.Helper()

	created, err := s.connections.Create(context.Background(),
		model_builders.NewMCPConnectionBuilder().WithAgentId(s.agentId).Build())
	assert.NoError(t, err)
	return created
}

func (s *unitOfWorkTestingSuite) nameOf(t *testing.T, id uuid.UUID) string {
	t.Helper()

	connection, err := s.connections.GetById(context.Background(), s.agentId, id)
	assert.NoError(t, err)
	return connection.Name
}
