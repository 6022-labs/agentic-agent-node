package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/clients"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_mcp/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_mcp_mocks/clients_mocks"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_mcp_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingMcpToolsThroughTheRequesterTestingSuite struct {
	sut *services.MCPRequester

	mockMcpClient     *clients_mocks.MockMcpClientInterface
	mockMcpMetrics    *metrics_mocks.MockMcpMetricsInterface
	testMcpConnection *models.MCPConnection
}

func WhenListingMcpToolsThroughTheRequesterBeforeEach(t *testing.T) *WhenListingMcpToolsThroughTheRequesterTestingSuite {
	mockController := gomock.NewController(t)

	mockMcpClient := clients_mocks.NewMockMcpClientInterface(mockController)
	mockMcpMetrics := metrics_mocks.NewMockMcpMetricsInterface(mockController)

	return &WhenListingMcpToolsThroughTheRequesterTestingSuite{
		sut: services.NewMCPRequester(mockMcpClient, mockMcpMetrics),

		mockMcpClient:  mockMcpClient,
		mockMcpMetrics: mockMcpMetrics,
		testMcpConnection: &models.MCPConnection{
			Name:       "github",
			Url:        "https://mcp.github.example",
			HeaderAuth: map[string]string{"Authorization": "Bearer wrong"},
		},
	}
}

func TestWhenListingMcpToolsThroughTheRequester(t *testing.T) {
	t.Parallel()

	t.Run("Given the server answers 401", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMcpToolsThroughTheRequesterTestingSuite) {
			unauthorized := &clients.McpStatusError{Method: "tools/list", StatusCode: http.StatusUnauthorized}
			suite.mockMcpClient.EXPECT().ListTools(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, unauthorized)
			suite.mockMcpMetrics.EXPECT().
				RecordListTools(gomock.Any(), suite.testMcpConnection, gomock.Any(), unauthorized, nil)
		}

		t.Run("Should surface the unauthorized sentinel", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsThroughTheRequesterBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListTools(context.Background(), suite.testMcpConnection)

			assert.ErrorIs(t, err, reasoning_interfaces.ErrMCPUnauthorized)
		})
	})

	t.Run("Given the server answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMcpToolsThroughTheRequesterTestingSuite) {
			serverError := &clients.McpStatusError{Method: "tools/list", StatusCode: http.StatusInternalServerError}
			suite.mockMcpClient.EXPECT().ListTools(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, serverError)
			suite.mockMcpMetrics.EXPECT().
				RecordListTools(gomock.Any(), suite.testMcpConnection, gomock.Any(), serverError, nil)
		}

		t.Run("Should not blame the credentials", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsThroughTheRequesterBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListTools(context.Background(), suite.testMcpConnection)

			assert.Error(t, err)
			assert.NotErrorIs(t, err, reasoning_interfaces.ErrMCPUnauthorized)

			var statusErr *clients.McpStatusError
			assert.ErrorAs(t, err, &statusErr)
			assert.Equal(t, http.StatusInternalServerError, statusErr.StatusCode)
		})
	})

	t.Run("Given the transport fails outright", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingMcpToolsThroughTheRequesterTestingSuite) {
			transportError := errors.New("connection refused")
			suite.mockMcpClient.EXPECT().ListTools(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, transportError)
			suite.mockMcpMetrics.EXPECT().
				RecordListTools(gomock.Any(), suite.testMcpConnection, gomock.Any(), transportError, nil)
		}

		t.Run("Should pass the reason through untouched", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingMcpToolsThroughTheRequesterBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ListTools(context.Background(), suite.testMcpConnection)

			assert.ErrorContains(t, err, "connection refused")
			assert.NotErrorIs(t, err, reasoning_interfaces.ErrMCPUnauthorized)
		})
	})
}
