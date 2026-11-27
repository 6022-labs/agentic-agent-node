package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_a2a_mocks/clients_mocks"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_a2a_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const fetchBaseUrl = "https://peer"

type WhenFetchingAgentCardTestingSuite struct {
	sut *services.A2ARequester

	mockA2AClient  *clients_mocks.MockA2AClientInterface
	mockA2AMetrics *metrics_mocks.MockA2AMetricsInterface
}

func WhenFetchingAgentCardBeforeEach(t *testing.T) *WhenFetchingAgentCardTestingSuite {
	mockController := gomock.NewController(t)

	mockA2AClient := clients_mocks.NewMockA2AClientInterface(mockController)
	mockA2AMetrics := metrics_mocks.NewMockA2AMetricsInterface(mockController)

	return &WhenFetchingAgentCardTestingSuite{
		sut:            services.NewA2ARequester(mockA2AClient, mockA2AMetrics),
		mockA2AClient:  mockA2AClient,
		mockA2AMetrics: mockA2AMetrics,
	}
}

func TestWhenFetchingAgentCard(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given the agent card cannot be fetched", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenFetchingAgentCardTestingSuite) {
			suite.mockA2AClient.EXPECT().FetchAgentCard(ctx, fetchBaseUrl).Return(nil, assert.AnError).Times(1)
			suite.mockA2AMetrics.EXPECT().
				RecordFetchAgentCard(ctx, fetchBaseUrl, gomock.Any(), assert.AnError).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.FetchAgentCard(ctx, fetchBaseUrl)

			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given the agent serves a card", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenFetchingAgentCardTestingSuite) {
			suite.mockA2AClient.EXPECT().
				FetchAgentCard(ctx, fetchBaseUrl).
				Return(&models.ExternalAgentCard{
					Name: "hermes",
					SupportedInterfaces: []models.ExternalAgentInterface{
						{Url: "https://peer/grpc", ProtocolBinding: "GRPC"},
						{Url: "https://peer/api/a2a", ProtocolBinding: models.JsonRpcProtocolBinding},
					},
				}, nil).
				Times(1)
			suite.mockA2AMetrics.EXPECT().RecordFetchAgentCard(ctx, fetchBaseUrl, gomock.Any(), nil).Times(1)
		}

		t.Run("Should map its jsonrpc interface onto the endpoint url", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t)
			initSuite(suite)

			card, err := suite.sut.FetchAgentCard(ctx, fetchBaseUrl)

			assert.NoError(t, err)
			assert.Equal(t, &interfaces.A2AAgentCard{Name: "hermes", EndpointUrl: "https://peer/api/a2a"}, card)
		})
	})

	t.Run("Given the agent card has no jsonrpc interface", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenFetchingAgentCardTestingSuite) {
			suite.mockA2AClient.EXPECT().
				FetchAgentCard(ctx, fetchBaseUrl).
				Return(&models.ExternalAgentCard{
					Name: "hermes",
					SupportedInterfaces: []models.ExternalAgentInterface{
						{Url: "https://peer/grpc", ProtocolBinding: "GRPC"},
					},
				}, nil).
				Times(1)
			suite.mockA2AMetrics.EXPECT().RecordFetchAgentCard(ctx, fetchBaseUrl, gomock.Any(), nil).Times(1)
		}

		t.Run("Should return no endpoint url", func(t *testing.T) {
			t.Parallel()

			suite := WhenFetchingAgentCardBeforeEach(t)
			initSuite(suite)

			card, err := suite.sut.FetchAgentCard(ctx, fetchBaseUrl)

			assert.NoError(t, err)
			assert.Empty(t, card.EndpointUrl)
		})
	})
}
