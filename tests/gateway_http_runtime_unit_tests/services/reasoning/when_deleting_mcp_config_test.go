package services_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDeletingMCPConfigTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingMCPConfigBeforeEach(t *testing.T) *WhenDeletingMCPConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingMCPConfigTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingMCPConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime deletes the connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).Return(&clients.DeleteInternalConfigMcpIdResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
			}, nil)
		}

		t.Run("Should not return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the runtime fails to delete the connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenDeletingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().DeleteInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).Return(&clients.DeleteInternalConfigMcpIdResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"boom"}`),
			}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingMCPConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background(), id)

			assert.Error(t, err)
		})
	})
}
