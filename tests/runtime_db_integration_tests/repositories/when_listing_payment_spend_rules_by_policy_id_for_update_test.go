package repositories_test

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	runtime_db_services "github.com/6022-labs/agentic-agent-node/src/runtime_db/services"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const listingForUpdateWindowSeconds = 24 * 60 * 60

type WhenListingPaymentSpendRulesByPolicyIdForUpdateTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleRepository
	unitOfWorkFactory        *runtime_db_services.UnitOfWorkFactory
	policy                   *models.PaymentSpendPolicy
}

func WhenListingPaymentSpendRulesByPolicyIdForUpdateBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingPaymentSpendRulesByPolicyIdForUpdateTestingSuite {
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
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentSpendPolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}

	return &WhenListingPaymentSpendRulesByPolicyIdForUpdateTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleRepository(dbClient),
		unitOfWorkFactory:        runtime_db_services.NewUnitOfWorkFactory(dbClient),
		policy:                   policy,
	}
}

func WhenListingPaymentSpendRulesByPolicyIdForUpdateAfterEach(
	suite *WhenListingPaymentSpendRulesByPolicyIdForUpdateTestingSuite,
) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingPaymentSpendRulesByPolicyIdForUpdate(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given concurrent scopes decide against one budget", func(t *testing.T) {
				t.Parallel()

				// The locked read is what stops both readers seeing the same empty period.
				t.Run("Should let through only what the budget funds", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentSpendRulesByPolicyIdForUpdateBeforeEach(tc)
					defer WhenListingPaymentSpendRulesByPolicyIdForUpdateAfterEach(suite)

					budgeted := model_builders.NewPaymentSpendRuleBuilder().
						WithPolicyId(suite.policy.Id).
						WithWindowSeconds(listingForUpdateWindowSeconds).
						WithMaxAmountPerWindow("300").
						Build()
					_, err := suite.sut.Create(context.Background(), budgeted)
					assert.NoError(t, err)

					const racers = 8
					results := make([]bool, racers)
					var waitGroup sync.WaitGroup
					waitGroup.Add(racers)
					for i := range racers {
						go func(index int) {
							defer waitGroup.Done()

							unitOfWork, err := suite.unitOfWorkFactory.Start(context.Background())
							assert.NoError(t, err)
							defer func() { _ = unitOfWork.Rollback() }()

							locked, err := suite.sut.ListByPolicyIdForUpdate(unitOfWork.Context(), suite.policy.Id)
							assert.NoError(t, err)

							recorded, err := locked[0].RecordWindowSpend(big.NewInt(100), time.Now())
							assert.NoError(t, err)
							if !recorded {
								return
							}
							_, err = suite.sut.Update(unitOfWork.Context(), &locked[0])
							assert.NoError(t, err)
							assert.NoError(t, unitOfWork.Commit())
							results[index] = true
						}(i)
					}
					waitGroup.Wait()

					granted := 0
					for _, allowed := range results {
						if allowed {
							granted++
						}
					}

					assert.Equal(t, 3, granted, "a budget of 300 must fund exactly three payments of 100")

					stored, err := suite.sut.GetById(context.Background(), budgeted.Id)
					assert.NoError(t, err)
					assert.Equal(t, "300", stored.WindowSpent)
				})
			})
		})
	}
}
