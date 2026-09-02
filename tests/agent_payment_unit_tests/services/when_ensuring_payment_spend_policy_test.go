package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	payment_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenEnsuringPaymentSpendPolicyTestingSuite struct {
	sut *services.PaymentSpendPolicyProvider

	mockAgentRepository              *identity_repositories_mocks.MockAgentRepositoryInterface
	mockPaymentSpendPolicyRepository *payment_repositories_mocks.MockPaymentSpendPolicyRepositoryInterface
	mockPaymentSpendRuleRepository   *payment_repositories_mocks.MockPaymentSpendRuleRepositoryInterface
	mockPaymentSpendMethodStrategies *payment_interfaces_mocks.MockPaymentSpendMethodStrategiesInterface
	mockPaymentSpendMethodStrategy   *payment_interfaces_mocks.MockPaymentSpendMethodStrategyInterface
}

func WhenEnsuringPaymentSpendPolicyBeforeEach(t *testing.T) *WhenEnsuringPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockPaymentSpendPolicyRepository := payment_repositories_mocks.NewMockPaymentSpendPolicyRepositoryInterface(
		mockController,
	)
	mockPaymentSpendRuleRepository := payment_repositories_mocks.NewMockPaymentSpendRuleRepositoryInterface(
		mockController,
	)
	mockPaymentSpendMethodStrategies := payment_interfaces_mocks.NewMockPaymentSpendMethodStrategiesInterface(
		mockController,
	)
	mockPaymentSpendMethodStrategy := payment_interfaces_mocks.NewMockPaymentSpendMethodStrategyInterface(
		mockController,
	)

	return &WhenEnsuringPaymentSpendPolicyTestingSuite{
		sut: services.NewPaymentSpendPolicyProvider(
			mockAgentRepository,
			mockPaymentSpendPolicyRepository,
			mockPaymentSpendRuleRepository,
			mockPaymentSpendMethodStrategies,
		),

		mockAgentRepository:              mockAgentRepository,
		mockPaymentSpendPolicyRepository: mockPaymentSpendPolicyRepository,
		mockPaymentSpendRuleRepository:   mockPaymentSpendRuleRepository,
		mockPaymentSpendMethodStrategies: mockPaymentSpendMethodStrategies,
		mockPaymentSpendMethodStrategy:   mockPaymentSpendMethodStrategy,
	}
}

func TestWhenEnsuringPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the agent already has a policy", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) *models.PaymentSpendPolicy {
			agent := model_builders.NewAgentBuilder().Build()
			existing := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				GetByAgentId(ctx, agent.Id).
				Return(existing, nil).
				Times(1)
			return existing
		}

		t.Run("Should return it without seeding a second rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			existing := initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.Equal(t, existing.Id, policy.Id)
		})
	})

	t.Run("Given the agent has no policy yet", func(t *testing.T) {
		t.Parallel()

		// The seeded rule carries an id of its own; the matcher must be keyed by
		// that rule, never by the policy that owns it.
		initSuite := func(
			suite *WhenEnsuringPaymentSpendPolicyTestingSuite,
		) (*models.PaymentSpendPolicy, *models.PaymentSpendRule, *json.RawMessage, **models.PaymentSpendRule) {
			agent := model_builders.NewAgentBuilder().Build()
			created := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()
			seededRule := model_builders.NewPaymentSpendRuleBuilder().WithPolicyId(created.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)

			seeded := new(*models.PaymentSpendRule)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(ctx, gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.PaymentSpendRule) (*models.PaymentSpendRule, error) {
					*seeded = rule
					seededRule.Name = rule.Name
					seededRule.MaxAmount = rule.MaxAmount
					seededRule.Method = rule.Method
					return seededRule, nil
				}).
				Times(1)
			suite.mockPaymentSpendMethodStrategies.EXPECT().
				For(enums.PaymentMethodX402).
				Return(suite.mockPaymentSpendMethodStrategy, true).
				Times(1)

			seededConfig := new(json.RawMessage)
			suite.mockPaymentSpendMethodStrategy.EXPECT().
				SaveConfig(ctx, seededRule.Id, gomock.Any()).
				DoAndReturn(func(_ context.Context, _ uuid.UUID, raw json.RawMessage) error {
					*seededConfig = raw
					return nil
				}).
				Times(1)

			return created, seededRule, seededConfig, seeded
		}

		t.Run("Should seed a catch-all rule so a fresh agent can pay", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			created, _, seededConfig, seeded := initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.Equal(t, created.Id, policy.Id)
			assert.Equal(t, enums.PaymentMethodX402, (*seeded).Method)

			var config models.PaymentSpendRuleX402Config
			assert.NoError(t, config.Parse(*seededConfig))
			assert.Equal(t, []string{".*"}, config.NetworkPatterns)
			assert.Equal(t, []string{".*"}, config.AssetPatterns)
			assert.Equal(t, []string{".*"}, config.PayToPatterns)
			assert.Equal(t, []string{".*"}, config.AssetTransferMethodPatterns)
		})

		t.Run("Should seed a cap that bounds nothing until an owner narrows it", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			_, _, _, seeded := initSuite(suite)

			_, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.Equal(
				t,
				"115792089237316195423570985008687907853269984665640564039457584007913129639935",
				(*seeded).MaxAmount,
			)
		})
	})

	// Unlike the receive side there is no mint gate: the wallet is bootstrapped
	// long before mint and the agent may need to pay first.
	t.Run("Given the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			created := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(ctx, gomock.Any()).
				Return(model_builders.NewPaymentSpendRuleBuilder().Build(), nil).
				Times(1)
			suite.mockPaymentSpendMethodStrategies.EXPECT().
				For(gomock.Any()).
				Return(suite.mockPaymentSpendMethodStrategy, true).
				Times(1)
			suite.mockPaymentSpendMethodStrategy.EXPECT().
				SaveConfig(ctx, gomock.Any(), gomock.Any()).
				Return(nil).
				Times(1)
		}

		t.Run("Should materialize the policy anyway", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.NotNil(t, policy)
		})
	})

	t.Run("Given a concurrent request won the insert race", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) *models.PaymentSpendPolicy {
			agent := model_builders.NewAgentBuilder().Build()
			winner := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, errors.New("duplicate key")).
				Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(winner, nil).Times(1)
			return winner
		}

		t.Run("Should return the row that won rather than fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			winner := initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.Equal(t, winner.Id, policy.Id)
		})
	})

	t.Run("Given the insert failed and the re-read finds nothing", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given reading the agent fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given the agent row does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(nil, nil).Times(1)
		}

		t.Run("Should fail rather than create an orphan policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given reading the existing policy fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().
				GetByAgentId(ctx, agent.Id).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given seeding the rule fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			created := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(ctx, gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given no strategy is registered for the seeded rule's method", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			created := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(ctx, gomock.Any()).
				Return(model_builders.NewPaymentSpendRuleBuilder().Build(), nil).
				Times(1)
			suite.mockPaymentSpendMethodStrategies.EXPECT().For(gomock.Any()).Return(nil, false).Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given persisting the seeded matcher fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			created := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			suite.mockPaymentSpendRuleRepository.EXPECT().
				Create(ctx, gomock.Any()).
				Return(model_builders.NewPaymentSpendRuleBuilder().Build(), nil).
				Times(1)
			suite.mockPaymentSpendMethodStrategies.EXPECT().
				For(gomock.Any()).
				Return(suite.mockPaymentSpendMethodStrategy, true).
				Times(1)
			suite.mockPaymentSpendMethodStrategy.EXPECT().
				SaveConfig(ctx, gomock.Any(), gomock.Any()).
				Return(assert.AnError).
				Times(1)
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})
}
