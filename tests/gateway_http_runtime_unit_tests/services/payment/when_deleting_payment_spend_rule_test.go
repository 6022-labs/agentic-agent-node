package services_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/payment"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentSpendRuleTestingSuite struct {
	sut *services.HttpPaymentSpendRuleClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingPaymentSpendRuleBeforeEach(t *testing.T) *WhenDeletingPaymentSpendRuleTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingPaymentSpendRuleTestingSuite{
		sut:        services.NewHttpPaymentSpendRuleClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingPaymentSpendRule(t *testing.T) {
	t.Parallel()

	t.Run("Given the rule is deleted", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(gomock.Any(), "r1").
				Return(&clients.DeleteInternalConfigPaymentSpendPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
				}, nil)
		}

		t.Run("Should delete without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			assert.NoError(t, suite.sut.Delete(context.Background(), "r1"))
		})
	})

	t.Run("Given the delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentSpendRuleTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigPaymentSpendPolicyRulesIdWithResponse(gomock.Any(), "r1").
				Return(&clients.DeleteInternalConfigPaymentSpendPolicyRulesIdResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentSpendRuleBeforeEach(t)
			initSuite(suite)

			assert.Error(t, suite.sut.Delete(context.Background(), "r1"))
		})
	})
}
