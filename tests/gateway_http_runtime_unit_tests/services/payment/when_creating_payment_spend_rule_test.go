package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/payment/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingPaymentSpendRuleTestingSuite struct {
	sut *services.HttpPaymentSpendRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingPaymentSpendRuleBeforeEach(t *testing.T) *WhenCreatingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingPaymentSpendRuleTestingSuite{
		sut:        services.NewHttpPaymentSpendRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentSpendPolicyRulesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, body io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigPaymentSpendPolicyRulesResponse, error) {
					var sent models.ExternalPaymentSpendRule
					assert.NoError(t, json.NewDecoder(body).Decode(&sent))
					assert.Equal(t, "Public price", sent.Name)
					return &clients.PostInternalConfigPaymentSpendPolicyRulesResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"id":"r1","name":"Public price"}`),
					}, nil
				})
		}

		t.Run("Should forward and decode the created rule", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Create(
				context.Background(),
				&models.ExternalPaymentSpendRule{Name: "Public price"},
			)

			assert.NoError(t, err)
			assert.Equal(t, "r1", rule.Id)
		})
	})

	t.Run("Given the runtime rejects the rule", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigPaymentSpendPolicyRulesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigPaymentSpendPolicyRulesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"asset","message":"no ERC-20 token answers"}`),
				}, nil)
		}

		t.Run("Should map the 400 to a validation error", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			rule, err := suite.sut.Create(context.Background(), &models.ExternalPaymentSpendRule{})

			assert.Nil(t, rule)
			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "asset", validationErr.Field)
		})
	})
}
