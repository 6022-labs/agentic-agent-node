package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingMCPConfigsOverHttpTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingMCPConfigsOverHttpBeforeEach(t *testing.T) *WhenListingMCPConfigsOverHttpTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingMCPConfigsOverHttpTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingMCPConfigsOverHttp(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the configured connections", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMCPConfigsOverHttpTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigMcpWithResponse(gomock.Any()).Return(&clients.GetInternalConfigMcpResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"connections":[{"id":"mcp-1","name":"tools","url":"https://mcp.example","headerAuthKeys":[]}]}`),
			}, nil)
		}

		t.Run("Should decode the MCP connection list payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMCPConfigsOverHttpBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.List(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) && assert.Len(t, payload.Connections, 1) {
				assert.Equal(t, "mcp-1", payload.Connections[0].Id)
				assert.Equal(t, "tools", payload.Connections[0].Name)
			}
		})
	})

	t.Run("Given listing fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMCPConfigsOverHttpTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigMcpWithResponse(gomock.Any()).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should propagate the transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMCPConfigsOverHttpBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background())

			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
