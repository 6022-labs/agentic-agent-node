package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type WhenListingPaymentReceiveRulesByPolicyIdTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentReceiveRuleRepository
	policy                   *models.PaymentReceivePolicy
}

func WhenListingPaymentReceiveRulesByPolicyIdBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenListingPaymentReceiveRulesByPolicyIdTestingSuite {
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

	policy, err := repositories.NewPaymentReceivePolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentReceivePolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}

	return &WhenListingPaymentReceiveRulesByPolicyIdTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentReceiveRuleRepository(dbClient),
		policy:                   policy,
	}
}

func WhenListingPaymentReceiveRulesByPolicyIdAfterEach(suite *WhenListingPaymentReceiveRulesByPolicyIdTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenListingPaymentReceiveRulesByPolicyId(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given no rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return an empty list", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentReceiveRulesByPolicyIdBeforeEach(tc)
					defer WhenListingPaymentReceiveRulesByPolicyIdAfterEach(suite)

					rules, err := suite.sut.ListByPolicyId(context.Background(), suite.policy.Id)

					assert.NoError(t, err)
					assert.Empty(t, rules)
				})
			})

			t.Run("Given several rules exist", func(t *testing.T) {
				t.Parallel()

				t.Run("Should return them oldest first", func(t *testing.T) {
					t.Parallel()
					suite := WhenListingPaymentReceiveRulesByPolicyIdBeforeEach(tc)
					defer WhenListingPaymentReceiveRulesByPolicyIdAfterEach(suite)

					// The builder dates rules at random, so the order is pinned here.
					first := model_builders.NewPaymentReceiveRuleBuilder().
						WithPolicyId(suite.policy.Id).
						WithName("first").
						Build()
					first.CreatedAt = time.Now().Add(-time.Hour)
					second := model_builders.NewPaymentReceiveRuleBuilder().
						WithPolicyId(suite.policy.Id).
						WithName("second").
						Build()
					second.CreatedAt = time.Now()
					_, err := suite.sut.Create(context.Background(), second)
					assert.NoError(t, err)
					_, err = suite.sut.Create(context.Background(), first)
					assert.NoError(t, err)

					rules, err := suite.sut.ListByPolicyId(context.Background(), suite.policy.Id)

					assert.NoError(t, err)
					if assert.Len(t, rules, 2) {
						assert.Equal(t, "first", rules[0].Name)
						assert.Equal(t, "second", rules[1].Name)
					}
				})
			})
		})
	}
}
