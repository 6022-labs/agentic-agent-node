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

type WhenDeletingLLMConfigOverHttpTestingSuite struct {
	sut *services.HttpLLMConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingLLMConfigOverHttpBeforeEach(t *testing.T) *WhenDeletingLLMConfigOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingLLMConfigOverHttpTestingSuite{
		sut:        services.NewHttpLLMConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingLLMConfigOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime deletes the connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigLlmWithResponse(gomock.Any()).Return(&clients.DeleteInternalConfigLlmResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
			}, nil)
		}

		t.Run("Should delete the LLM connection without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConfigOverHttpBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			assert.NoError(t, err)
		})
	})

	t.Run("Given the delete fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigLlmWithResponse(gomock.Any()).Return(&clients.DeleteInternalConfigLlmResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"boom"}`),
			}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConfigOverHttpBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			assert.Error(t, err)
		})
	})
}
