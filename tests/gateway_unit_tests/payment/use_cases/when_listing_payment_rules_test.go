package use_cases_test

import (
	"context"
	"errors"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingPaymentRulesTestingSuite struct {
	sut *use_cases.ListPaymentRules

	mockPaymentRuleClient *interfaces_mocks.MockPaymentRuleClientInterface
}

func WhenListingPaymentRulesBeforeEach(t *testing.T) *WhenListingPaymentRulesTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentRuleClient := interfaces_mocks.NewMockPaymentRuleClientInterface(mockController)

	return &WhenListingPaymentRulesTestingSuite{
		sut: use_cases.NewListPaymentRules(mockPaymentRuleClient),

		mockPaymentRuleClient: mockPaymentRuleClient,
	}
}

func TestWhenListingPaymentRules(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns rules", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return([]models.ExternalPaymentRule{{Id: "rule-id", Name: "Public price"}}, nil).
				Times(1)
		}

		t.Run("Should return them", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, result.Rules, 1) {
				assert.Equal(t, "rule-id", result.Rules[0].Id)
			}
		})
	})

	t.Run("Given the runtime is unavailable", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingPaymentRulesTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				List(gomock.Any()).
				Return(nil, errors.New("connection refused")).
				Times(1)
		}

		t.Run("Should return an unavailable error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingPaymentRulesBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Execute(context.Background())

			assert.Nil(t, result)
			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})
}
