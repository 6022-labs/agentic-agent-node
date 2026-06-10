package services_test

import (
	"context"
	"net/http"
	"testing"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingMCPConfigTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingMCPConfigBeforeEach(t *testing.T) *WhenGettingMCPConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingMCPConfigTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingMCPConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the requested connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenGettingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).Return(&clients.GetInternalConfigMcpIdResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"id":"mcp-1","name":"tools","url":"https://mcp.example","headerAuthKeys":[]}`),
			}, nil)
		}

		t.Run("Should decode the MCP connection payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConfigBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.GetById(context.Background(), id)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "tools", payload.Name)
			}
		})
	})

	t.Run("Given the connection id does not exist", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenGettingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigMcpIdWithResponse(gomock.Any(), id.String()).Return(&clients.GetInternalConfigMcpIdResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
				Body:         []byte(`{"code":"not_found"}`),
			}, nil)
		}

		t.Run("Should return ErrMCPConnectionNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.GetById(context.Background(), id)

			assert.ErrorIs(t, err, gateway_interfaces.ErrMCPConnectionNotFound)
		})
	})
}
