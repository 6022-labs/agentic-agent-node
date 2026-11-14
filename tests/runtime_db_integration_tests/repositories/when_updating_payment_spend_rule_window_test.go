package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/repositories"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/seeds"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type WhenUpdatingPaymentSpendRuleWindowTestingSuite struct {
	dbName                   string
	dbClient                 *runtime_db.DatabaseClient
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	sut                      *repositories.PaymentSpendRuleWindowRepository
	rule                     *models.PaymentSpendRule
}

func WhenUpdatingPaymentSpendRuleWindowBeforeEach(
	tc *common_tests.DatabaseTestCase,
) *WhenUpdatingPaymentSpendRuleWindowTestingSuite {
	databaseContainerBuilder, dbName, dbClient := newMigratedDatabase(tc)

	agentId := uuid.New()
	seeds.SeedAgent(dbClient, agentId)

	policy, err := repositories.NewPaymentSpendPolicyRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agentId).Build())
	if err != nil {
		panic(err)
	}
	rule, err := repositories.NewPaymentSpendRuleRepository(dbClient).Create(
		context.Background(), model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(policy.Id).Build())
	if err != nil {
		panic(err)
	}

	return &WhenUpdatingPaymentSpendRuleWindowTestingSuite{
		dbName:                   dbName,
		dbClient:                 dbClient,
		databaseContainerBuilder: databaseContainerBuilder,
		sut:                      repositories.NewPaymentSpendRuleWindowRepository(dbClient),
		rule:                     rule,
	}
}

func WhenUpdatingPaymentSpendRuleWindowAfterEach(suite *WhenUpdatingPaymentSpendRuleWindowTestingSuite) {
	_ = suite.dbClient.Close()
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func TestWhenUpdatingPaymentSpendRuleWindow(t *testing.T) {
	t.Parallel()

	for _, tc := range common_tests.DatabaseTestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given the window exists", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store its budget and running spend at full width", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentSpendRuleWindowBeforeEach(tc)
					defer WhenUpdatingPaymentSpendRuleWindowAfterEach(suite)

					window, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().WithPaymentSpendRuleId(suite.rule.Id).Build())
					assert.NoError(t, err)

					maxUint256 := "115792089237316195423570985008687907853269984665640564039457584007913129639935"
					startedAt := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
					window.MaxAmount = maxUint256
					window.Spent = maxUint256
					window.StartedAt = &startedAt
					_, err = suite.sut.Update(context.Background(), window)
					assert.NoError(t, err)

					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					if assert.Len(t, windows, 1) {
						assert.Equal(t, window.Id, windows[0].Id)
						assert.Equal(t, maxUint256, windows[0].MaxAmount)
						assert.Equal(t, maxUint256, windows[0].Spent)
						if assert.NotNil(t, windows[0].StartedAt) {
							assert.True(t, startedAt.Equal(*windows[0].StartedAt))
						}
					}
				})
			})

			t.Run("Given a started window is reset", func(t *testing.T) {
				t.Parallel()

				t.Run("Should store no start and nothing spent", func(t *testing.T) {
					t.Parallel()
					suite := WhenUpdatingPaymentSpendRuleWindowBeforeEach(tc)
					defer WhenUpdatingPaymentSpendRuleWindowAfterEach(suite)

					window, err := suite.sut.Create(context.Background(),
						model_builders.NewPaymentSpendRuleWindowBuilder().
							WithPaymentSpendRuleId(suite.rule.Id).
							WithSpent("4000").
							WithStartedAt(time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)).
							Build())
					assert.NoError(t, err)

					window.Reset()
					_, err = suite.sut.Update(context.Background(), window)
					assert.NoError(t, err)

					windows, err := suite.sut.ListByRuleId(context.Background(), suite.rule.Id)
					assert.NoError(t, err)
					if assert.Len(t, windows, 1) {
						assert.Equal(t, "0", windows[0].Spent)
						assert.Nil(t, windows[0].StartedAt)
					}
				})
			})
		})
	}
}
