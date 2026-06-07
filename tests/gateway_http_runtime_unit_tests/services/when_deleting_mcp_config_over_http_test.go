package services_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingMCPConfigOverHttpTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingMCPConfigOverHttpBeforeEach(t *testing.T) *WhenDeletingMCPConfigOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingMCPConfigOverHttpTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingMCPConfigOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime deletes the connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).Return(&clients.DeleteInternalConfigMcpIdResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
			}, nil)
		}

		t.Run("Should not return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigOverHttpBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the runtime fails to delete the connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigOverHttpTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).Return(&clients.DeleteInternalConfigMcpIdResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"boom"}`),
			}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigOverHttpBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			assert.Error(t, err)
		})
	})
}
