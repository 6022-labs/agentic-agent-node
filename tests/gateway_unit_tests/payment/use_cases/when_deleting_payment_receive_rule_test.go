package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentReceiveRuleTestingSuite struct {
	sut *use_cases.DeletePaymentReceiveRule

	mockPaymentReceiveRuleClient *interfaces_mocks.MockPaymentReceiveRuleClientInterface
}

func WhenDeletingPaymentReceiveRuleBeforeEach(t *testing.T) *WhenDeletingPaymentReceiveRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentReceiveRuleClient := interfaces_mocks.NewMockPaymentReceiveRuleClientInterface(mockController)

	return &WhenDeletingPaymentReceiveRuleTestingSuite{
		sut: use_cases.NewDeletePaymentReceiveRule(mockPaymentReceiveRuleClient),

		mockPaymentReceiveRuleClient: mockPaymentReceiveRuleClient,
	}
}

func TestWhenDeletingPaymentReceiveRule(t *testing.T) {
	t.Parallel()

	ruleId := uuid.New()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewUnavailableError("runtime_unavailable", "runtime down", nil))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentReceiveRuleRequest{Id: ruleId})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewValidationError("id", "id is invalid"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentReceiveRuleRequest{Id: ruleId})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "id", validationErr.Field)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewNotFoundError("rule_not_found", "rule not found"))
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentReceiveRuleRequest{Id: ruleId})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "rule_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().
				Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewUnavailableError("runtime_unavailable", "runtime down", nil)).
				Times(1)
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentReceiveRuleRequest{Id: ruleId})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the rule is deleted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentReceiveRuleTestingSuite) {
			suite.mockPaymentReceiveRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).Return(nil).Times(1)
		}

		t.Run("Should delete without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentReceiveRuleBeforeEach(t)
			initSuite(suite)

			assert.NoError(
				t,
				suite.sut.Execute(context.Background(), &requests.DeletePaymentReceiveRuleRequest{Id: ruleId}),
			)
		})
	})
}
