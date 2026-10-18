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

type WhenDeletingPaymentRuleTestingSuite struct {
	sut *use_cases.DeletePaymentRule

	mockPaymentRuleClient *interfaces_mocks.MockPaymentRuleClientInterface
}

func WhenDeletingPaymentRuleBeforeEach(t *testing.T) *WhenDeletingPaymentRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentRuleClient := interfaces_mocks.NewMockPaymentRuleClientInterface(mockController)

	return &WhenDeletingPaymentRuleTestingSuite{
		sut: use_cases.NewDeletePaymentRule(mockPaymentRuleClient),

		mockPaymentRuleClient: mockPaymentRuleClient,
	}
}

func TestWhenDeletingPaymentRule(t *testing.T) {
	t.Parallel()

	ruleId := uuid.New()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewUnavailableError("runtime_unavailable", "runtime down"))
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: ruleId})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewValidationError("id", "id is invalid"))
		}

		t.Run("Should surface the ValidationError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: ruleId})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "id", validationErr.Field)
		})
	})

	t.Run("Given the rule does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewNotFoundError("rule_not_found", "rule not found"))
		}

		t.Run("Should surface the NotFoundError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: ruleId})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "rule_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				Delete(gomock.Any(), ruleId.String()).
				Return(apperrors.NewUnavailableError("runtime_unavailable", "runtime down")).
				Times(1)
		}

		t.Run("Should surface the UnavailableError unchanged", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: ruleId})

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
		})
	})

	t.Run("Given the rule is deleted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().Delete(gomock.Any(), ruleId.String()).Return(nil).Times(1)
		}

		t.Run("Should delete without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			initSuite(suite)

			assert.NoError(t, suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: ruleId}))
		})
	})
}
