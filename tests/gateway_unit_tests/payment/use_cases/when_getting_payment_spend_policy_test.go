package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingPaymentSpendPolicyTestingSuite struct {
	sut *use_cases.GetPaymentSpendPolicy

	mockPaymentSpendPolicyClient *interfaces_mocks.MockPaymentSpendPolicyClientInterface
}

func WhenGettingPaymentSpendPolicyBeforeEach(t *testing.T) *WhenGettingPaymentSpendPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendPolicyClient := interfaces_mocks.NewMockPaymentSpendPolicyClientInterface(mockController)

	return &WhenGettingPaymentSpendPolicyTestingSuite{
		sut: use_cases.NewGetPaymentSpendPolicy(mockPaymentSpendPolicyClient),

		mockPaymentSpendPolicyClient: mockPaymentSpendPolicyClient,
	}
}

func TestWhenGettingPaymentSpendPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the policy head", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Get(gomock.Any()).
				Return(&models.ExternalPaymentSpendPolicyPayload{
					Id:            "policy-id",
					AgentId:       "agent-id",
					PayToPatterns: []string{".*"},
				}, nil).
				Times(1)
		}

		t.Run("Should return the head", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "policy-id", result.Id)
			assert.Equal(t, []string{".*"}, result.PayToPatterns)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Get(gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "connection refused")).
				Times(1)
		}

		t.Run("Should return an unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentSpendPolicyTestingSuite) {
			suite.mockPaymentSpendPolicyClient.EXPECT().
				Get(gomock.Any()).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "agent identity is not minted yet")).
				Times(1)
		}

		t.Run("Should pass the conflict through", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentSpendPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})
}
