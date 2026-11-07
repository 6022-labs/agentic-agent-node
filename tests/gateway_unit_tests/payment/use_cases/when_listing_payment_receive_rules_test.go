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

type WhenListingPaymentReceiveRulesTestingSuite struct {
	sut *use_cases.ListPaymentReceiveRules

	mockPaymentReceiveRuleClient *interfaces_mocks.MockPaymentReceiveRuleClientInterface
}

func WhenListingPaymentReceiveRulesBeforeEach(t *testing.T) *WhenListingPaymentReceiveRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceiveRuleClient := interfaces_mocks.NewMockPaymentReceiveRuleClientInterface(mockController)

	return &WhenListingPaymentReceiveRulesTestingSuite{
		sut: use_cases.NewListPaymentReceiveRules(mockPaymentReceiveRuleClient),

		mockPaymentReceiveRuleClient: mockPaymentReceiveRuleClient,
	}
}

func TestWhenListingPaymentReceiveRules(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(gomock.Any()).
				Return(nil, apperrors.NewUnavailableError("runtime_unavailable", "runtime down", nil))
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime refuses because the agent is not minted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(gomock.Any()).
				Return(nil, apperrors.NewConflictError("agent_not_minted", "agent identity is not minted yet")).
				Times(1)
		}

		t.Run("Should surface the ConflictError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "agent_not_minted", conflictErr.Code)
		})
	})

	t.Run("Given the runtime lists rules", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentReceiveRulesTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				List(gomock.Any()).
				Return([]models.ExternalPaymentReceiveRule{{Id: "rule-id", Name: "Public price"}}, nil).
				Times(1)
		}

		t.Run("Should return them", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentReceiveRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, result.Rules, 1) {
				assert.Equal(t, "rule-id", result.Rules[0].Id)
			}
		})
	})
}
