package services_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingPaymentPolicyTestingSuite struct {
	sut *services.HttpPaymentPolicyClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingPaymentPolicyBeforeEach(t *testing.T) *WhenDeletingPaymentPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingPaymentPolicyTestingSuite{
		sut:        services.NewHttpPaymentPolicyClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingPaymentPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime deletes the configuration", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentPolicyTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigPaymentPolicyWithResponse(gomock.Any()).Return(&clients.DeleteInternalConfigPaymentPolicyResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
			}, nil)
		}

		t.Run("Should delete the configuration without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingPaymentPolicyTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigPaymentPolicyWithResponse(gomock.Any()).Return(&clients.DeleteInternalConfigPaymentPolicyResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"boom"}`),
			}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			assert.Error(t, err)
		})
	})
}
