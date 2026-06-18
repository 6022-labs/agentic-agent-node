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

type WhenGettingPaymentPolicyTestingSuite struct {
	sut *services.HttpPaymentPolicyClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingPaymentPolicyBeforeEach(t *testing.T) *WhenGettingPaymentPolicyTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingPaymentPolicyTestingSuite{
		sut:        services.NewHttpPaymentPolicyClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingPaymentPolicy(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the policy head", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigPaymentPolicyWithResponse(gomock.Any()).Return(&clients.GetInternalConfigPaymentPolicyResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"id":"a","agentId":"b","agentWalletAddress":"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"}`),
			}, nil)
		}

		t.Run("Should decode the head", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "a", payload.Id)
			assert.Equal(t, "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266", payload.AgentWalletAddress)
		})
	})

	t.Run("Given the runtime fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingPaymentPolicyTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigPaymentPolicyWithResponse(gomock.Any()).Return(&clients.GetInternalConfigPaymentPolicyResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"get_failed"}`),
			}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingPaymentPolicyBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.Nil(t, payload)
			assert.Error(t, err)
		})
	})
}
