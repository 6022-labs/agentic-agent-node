package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingPaymentSpendRuleTestingSuite struct {
	sut *use_cases.CreatePaymentSpendRule

	mockPaymentSpendRuleClient *interfaces_mocks.MockPaymentSpendRuleClientInterface
}

func WhenCreatingPaymentSpendRuleBeforeEach(t *testing.T) *WhenCreatingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentSpendRuleClient := interfaces_mocks.NewMockPaymentSpendRuleClientInterface(mockController)

	return &WhenCreatingPaymentSpendRuleTestingSuite{
		sut: use_cases.NewCreatePaymentSpendRule(mockPaymentSpendRuleClient),

		mockPaymentSpendRuleClient: mockPaymentSpendRuleClient,
	}
}

func TestWhenCreatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given a rule to create", func(t *testing.T) {
		t.Parallel()

		// The gateway forwards verbatim; the runtime owns validation.
		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			suite.mockPaymentSpendRuleClient.EXPECT().
				Create(gomock.Any(), &models.ExternalPaymentSpendRule{
					Name:                "Only base, small",
					MaxAmountPerPayment: lo.ToPtr("1000"),
					Windows: []models.ExternalPaymentSpendRuleWindow{
						{WindowSeconds: 86400, MaxAmount: "50000"},
						{WindowSeconds: 604800, MaxAmount: "200000"},
					},
					Method: "x402",
					Config: json.RawMessage(`{"network":"eip155:8453"}`),
				}).
				Return(&models.ExternalPaymentSpendRule{Id: "rule-id", Name: "Only base, small"}, nil).
				Times(1)
		}

		t.Run("Should forward every field to the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.SavePaymentSpendRuleRequest{
				Name:                "Only base, small",
				MaxAmountPerPayment: lo.ToPtr("1000"),
				Windows: []requests.SavePaymentSpendRuleWindowRequest{
					{WindowSeconds: 86400, MaxAmount: "50000"},
					{WindowSeconds: 604800, MaxAmount: "200000"},
				},
				Method: "x402",
				Config: json.RawMessage(`{"network":"eip155:8453"}`),
			})

			assert.NoError(t, err)
			assert.Equal(t, "rule-id", rule.Id)
		})
	})

	t.Run("Given the runtime rejects the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			suite.mockPaymentSpendRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, assert.AnError).
				Times(1)
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.SavePaymentSpendRuleRequest{})

			assert.Nil(t, rule)
			assert.ErrorIs(t, err, assert.AnError)
		})
	})
}
