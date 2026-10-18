package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingMCPConfigsTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenListingMCPConfigsBeforeEach(t *testing.T) *WhenListingMCPConfigsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenListingMCPConfigsTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenListingMCPConfigs(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMCPConfigsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigMcpWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMCPConfigsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMCPConfigsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigMcpWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigMcpResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMCPConfigsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.List(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime returns the configured connections", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMCPConfigsTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalConfigMcpWithResponse(gomock.Any()).
				Return(&clients.GetInternalConfigMcpResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body: []byte(
						`{"connections":[{"id":"mcp-1","name":"tools","url":"https://mcp.example","headerAuthKeys":[]}]}`,
					),
				}, nil)
		}

		t.Run("Should decode the MCP connection list payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMCPConfigsBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.List(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) && assert.Len(t, payload.Connections, 1) {
				assert.Equal(t, "mcp-1", payload.Connections[0].Id)
				assert.Equal(t, "tools", payload.Connections[0].Name)
			}
		})
	})
}
