package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_payment/services"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	payment_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_payment_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenEnsuringPaymentSpendPolicyTestingSuite struct {
	sut *services.PaymentSpendPolicyProvider

	mockAgentRepository              *identity_repositories_mocks.MockAgentRepositoryInterface
	mockPaymentSpendPolicyRepository *payment_repositories_mocks.MockPaymentSpendPolicyRepositoryInterface
}

func WhenEnsuringPaymentSpendPolicyBeforeEach(t *testing.T) *WhenEnsuringPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)
	mockPaymentSpendPolicyRepository := payment_repositories_mocks.NewMockPaymentSpendPolicyRepositoryInterface(
		mockController,
	)

	return &WhenEnsuringPaymentSpendPolicyTestingSuite{
		sut: services.NewPaymentSpendPolicyProvider(
			mockAgentRepository,
			mockPaymentSpendPolicyRepository,
		),

		mockAgentRepository:              mockAgentRepository,
		mockPaymentSpendPolicyRepository: mockPaymentSpendPolicyRepository,
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
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(existing, nil).Times(1)
			return existing
		}

		t.Run("Should return it", func(t *testing.T) {
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

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) *models.PaymentSpendPolicy {
			agent := model_builders.NewAgentBuilder().Build()
			created := model_builders.NewPaymentSpendPolicyBuilder().WithAgentId(agent.Id).Build()

			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().GetByAgentId(ctx, agent.Id).Return(nil, nil).Times(1)
			suite.mockPaymentSpendPolicyRepository.EXPECT().Create(ctx, gomock.Any()).Return(created, nil).Times(1)
			return created
		}

		// Deny by default: no rule is seeded, so the agent spends nothing until an owner writes one.
		t.Run("Should create an empty policy without seeding a rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			created := initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.NoError(t, err)
			assert.Equal(t, created.Id, policy.Id)
		})
	})

	t.Run("Given the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithoutCollectionTokenId().Build()
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(agent, nil).Times(1)
		}

		t.Run("Should return a conflict without touching the policy", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var conflictError *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictError)
			assert.Equal(t, "agent_not_minted", conflictError.Code)
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

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})

	t.Run("Given reading the agent fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenEnsuringPaymentSpendPolicyTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(ctx).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
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
			assert.Equal(t, "spend_policy_failed", internalError.Code)
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

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenEnsuringPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			policy, err := suite.sut.Ensure(ctx)

			assert.Nil(t, policy)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
			assert.Equal(t, "spend_policy_failed", internalError.Code)
		})
	})
}
