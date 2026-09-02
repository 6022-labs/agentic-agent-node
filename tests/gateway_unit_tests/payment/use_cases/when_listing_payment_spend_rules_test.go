package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingPaymentSpendRulesTestingSuite struct {
	sut *use_cases.ListPaymentSpendRules

	mockPaymentSpendRuleClient *interfaces_mocks.MockPaymentSpendRuleClientInterface
}

func WhenListingPaymentSpendRulesBeforeEach(t *testing.T) *WhenListingPaymentSpendRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleClient := interfaces_mocks.NewMockPaymentSpendRuleClientInterface(mockController)

	return &WhenListingPaymentSpendRulesTestingSuite{
		sut: use_cases.NewListPaymentSpendRules(mockPaymentSpendRuleClient),

		mockPaymentSpendRuleClient: mockPaymentSpendRuleClient,
	}
}

func TestWhenListingPaymentSpendRules(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns rules", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockPaymentSpendRuleClient.EXPECT().
				List(gomock.Any()).
				Return([]models.ExternalPaymentSpendRule{{Id: "rule-id", Name: "Anything, unbounded"}}, nil).
				Times(1)
		}

		t.Run("Should return them in table order", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Len(t, response.Rules, 1)
			assert.Equal(t, "Anything, unbounded", response.Rules[0].Name)
		})
	})

	t.Run("Given the runtime fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentSpendRulesTestingSuite) {
			suite.mockPaymentSpendRuleClient.EXPECT().List(gomock.Any()).Return(nil, assert.AnError).Times(1)
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentSpendRulesBeforeEach(t)
			initSuite(suite)

			response, err := suite.sut.Execute(context.Background())

			assert.Nil(t, response)
			assert.Error(t, err)
		})
	})
}
