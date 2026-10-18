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

type WhenGettingPaymentPolicyTestingSuite struct {
	sut *use_cases.GetPaymentPolicy

	mockPaymentPolicyClient *interfaces_mocks.MockPaymentPolicyClientInterface
}

func WhenGettingPaymentPolicyBeforeEach(t *testing.T) *WhenGettingPaymentPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentPolicyClient := interfaces_mocks.NewMockPaymentPolicyClientInterface(mockController)

	return &WhenGettingPaymentPolicyTestingSuite{
		sut: use_cases.NewGetPaymentPolicy(mockPaymentPolicyClient),

		mockPaymentPolicyClient: mockPaymentPolicyClient,
	}
}

func TestWhenGettingPaymentPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) {
			suite.mockPaymentPolicyClient.EXPECT().
				Get(gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) {
			suite.mockPaymentPolicyClient.EXPECT().
				Get(gomock.Any()).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "agent identity is not minted yet")).
				Times(1)
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime returns the policy head", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) {
			suite.mockPaymentPolicyClient.EXPECT().
				Get(gomock.Any()).
				Return(&models.ExternalPaymentPolicyPayload{
					Id:                 "policy-id",
					AgentWalletAddress: "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
				}, nil).
				Times(1)
		}

		t.Run("Should return the head", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "policy-id", result.Id)
			assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", result.AgentWalletAddress)
		})
	})
}
