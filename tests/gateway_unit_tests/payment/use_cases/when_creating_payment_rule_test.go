package use_cases_test

import (
	"context"
	"encoding/json"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/payment_mocks/services_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingPaymentRuleTestingSuite struct {
	sut *use_cases.CreatePaymentRule

	mockPaymentRuleClient *interfaces_mocks.MockPaymentRuleClientInterface
}

func WhenCreatingPaymentRuleBeforeEach(t *testing.T) *WhenCreatingPaymentRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockPaymentRuleClient := interfaces_mocks.NewMockPaymentRuleClientInterface(mockController)

	return &WhenCreatingPaymentRuleTestingSuite{
		sut: use_cases.NewCreatePaymentRule(mockPaymentRuleClient),

		mockPaymentRuleClient: mockPaymentRuleClient,
	}
}

func TestWhenCreatingPaymentRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, rule *models.ExternalPaymentRule) (*models.ExternalPaymentRule, error) {
					assert.Equal(t, "Public price", rule.Name)
					assert.Equal(t, "x402", rule.Method)
					assert.JSONEq(t, `{"network":"eip155:80002"}`, string(rule.Config))
					rule.Id = "rule-id"
					return rule, nil
				}).
				Times(1)
		}

		t.Run("Should forward the rule and return it", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.SavePaymentRuleRequest{
				Name:   "Public price",
				Method: "x402",
				Config: json.RawMessage(`{"network":"eip155:80002"}`),
			})

			assert.NoError(t, err)
			assert.Equal(t, "rule-id", rule.Id)
		})
	})

	t.Run("Given the runtime rejects the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentRuleTestingSuite) {
			suite.mockPaymentRuleClient.EXPECT().
				Create(gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewValidationError("asset", "no ERC-20 token answers")).
				Times(1)
		}

		t.Run("Should pass the validation error through", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Execute(context.Background(), &requests.SavePaymentRuleRequest{})

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})
}
