package use_cases_test

import (
	"context"
	"errors"
	"testing"

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

	t.Run("Given the delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				Delete(gomock.Any(), ruleId.String()).
				Return(errors.New("runtime down")).
				Times(1)
		}

		t.Run("Should pass the error through", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentRuleBeforeEach(t)
			initSuite(suite)

			assert.Error(t, suite.sut.Execute(context.Background(), &requests.DeletePaymentRuleRequest{Id: ruleId}))
		})
	})
}
