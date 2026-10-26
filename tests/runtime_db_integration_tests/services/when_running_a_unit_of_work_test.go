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

type WhenRunningAUnitOfWorkTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *runtime_db_services.UnitOfWorkFactory
	connections              *repositories.MCPConnectionRepository
	agentId                  uuid.UUID
}

func WhenRunningAUnitOfWorkBeforeEach(tc *common_tests.DatabaseTestCase) *WhenRunningAUnitOfWorkTestingSuite {
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

	return &WhenRunningAUnitOfWorkTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      runtime_db_services.NewUnitOfWorkFactory(dbClient),
		connections:              repositories.NewMCPConnectionRepository(dbClient),
		agentId:                  agentId,
	}
}

func WhenRunningAUnitOfWorkAfterEach(suite *WhenRunningAUnitOfWorkTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func (s *WhenRunningAUnitOfWorkTestingSuite) seedConnection(t *testing.T) *models.MCPConnection {
	t.Helper()

	created, err := s.connections.Create(context.Background(),
		model_builders.NewMCPConnectionBuilder().WithAgentId(s.agentId).Build())
	assert.NoError(t, err)
	return created
}

func (s *WhenRunningAUnitOfWorkTestingSuite) nameOf(t *testing.T, id uuid.UUID) string {
	t.Helper()

	connection, err := s.connections.GetById(context.Background(), s.agentId, id)
	assert.NoError(t, err)
	return connection.Name
}

func TestWhenRunningAUnitOfWork(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the work is committed", func(t *testing.T) {
				t.Parallel()

				t.Run("Should keep every write made through the scope", func(t *testing.T) {
					t.Parallel()
					suite := WhenRunningAUnitOfWorkBeforeEach(tc)
					defer WhenRunningAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "kept"
					_, err = suite.connections.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Commit())

					assert.Equal(t, "kept", suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given the scope is rolled back", func(t *testing.T) {
				t.Parallel()

				t.Run("Should discard every write made through it", func(t *testing.T) {
					t.Parallel()
					suite := WhenRunningAUnitOfWorkBeforeEach(tc)
					defer WhenRunningAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "discarded"
					_, err = suite.connections.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, unitOfWork.Rollback())

					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given the caller never commits", func(t *testing.T) {
				t.Parallel()

				// The pattern every caller uses: defer Rollback, return early on refusal.
				t.Run("Should discard the work when the deferred rollback runs", func(t *testing.T) {
					t.Parallel()
					suite := WhenRunningAUnitOfWorkBeforeEach(tc)
					defer WhenRunningAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					func() {
						unitOfWork, err := suite.sut.Start(context.Background())
						assert.NoError(t, err)
						defer func() { _ = unitOfWork.Rollback() }()

						connection.Name = "abandoned"
						_, err = suite.connections.Update(unitOfWork.Context(), connection)
						assert.NoError(t, err)
					}()

					assert.Equal(t, original, suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given a rollback follows a commit", func(t *testing.T) {
				t.Parallel()

				// `defer Rollback()` sits on the committed path too, so it must be a no-op.
				t.Run("Should be a no-op that keeps the committed work", func(t *testing.T) {
					t.Parallel()
					suite := WhenRunningAUnitOfWorkBeforeEach(tc)
					defer WhenRunningAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)

					unitOfWork, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)
					connection.Name = "kept"
					_, err = suite.connections.Update(unitOfWork.Context(), connection)
					assert.NoError(t, err)

					assert.NoError(t, unitOfWork.Commit())
					assert.NoError(t, unitOfWork.Rollback())

					assert.Equal(t, "kept", suite.nameOf(t, connection.Id))
				})
			})

			t.Run("Given a nested scope is started", func(t *testing.T) {
				t.Parallel()

				// A nested caller must not be able to keep or discard work it does not own.
				t.Run("Should join the outer scope rather than commit on its own", func(t *testing.T) {
					t.Parallel()
					suite := WhenRunningAUnitOfWorkBeforeEach(tc)
					defer WhenRunningAUnitOfWorkAfterEach(suite)

					connection := suite.seedConnection(t)
					original := connection.Name

					outer, err := suite.sut.Start(context.Background())
					assert.NoError(t, err)

					inner, err := suite.sut.Start(outer.Context())
					assert.NoError(t, err)
					connection.Name = "written by the inner scope"
					_, err = suite.connections.Update(inner.Context(), connection)
					assert.NoError(t, err)
					assert.NoError(t, inner.Commit())

					assert.NoError(t, outer.Rollback())
					assert.Equal(t, original, suite.nameOf(t, connection.Id),
						"the inner commit must not have escaped the outer rollback")
				})
			})
		})
	}
}
