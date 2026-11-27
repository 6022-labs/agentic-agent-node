package services_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_http_a2a/services"
	"github.com/6022-labs/agentic-agent-node/src/common/x402"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_a2a_mocks/clients_mocks"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_http_a2a_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	sendEndpointUrl = "https://peer/api/a2a"
	sendMessage     = "hello"
	sendContextId   = "conversation-1"
	sendSignature   = "eyJ4NDAy"
)

type WhenSendingA2AMessageTestingSuite struct {
	sut *services.A2ARequester

	mockA2AClient  *clients_mocks.MockA2AClientInterface
	mockA2AMetrics *metrics_mocks.MockA2AMetricsInterface
}

func WhenSendingA2AMessageBeforeEach(t *testing.T) *WhenSendingA2AMessageTestingSuite {
	mockController := gomock.NewController(t)

	mockA2AClient := clients_mocks.NewMockA2AClientInterface(mockController)
	mockA2AMetrics := metrics_mocks.NewMockA2AMetricsInterface(mockController)

	return &WhenSendingA2AMessageTestingSuite{
		sut:            services.NewA2ARequester(mockA2AClient, mockA2AMetrics),
		mockA2AClient:  mockA2AClient,
		mockA2AMetrics: mockA2AMetrics,
	}
}

func (suite *WhenSendingA2AMessageTestingSuite) send() (*interfaces.A2AReply, *x402.PaymentRequired, error) {
	return suite.sut.SendMessage(context.Background(), sendEndpointUrl, sendMessage, sendContextId, sendSignature)
}

func TestWhenSendingA2AMessage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given sending the message fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockA2AClient.EXPECT().
				SendMessage(ctx, sendEndpointUrl, gomock.Any(), sendSignature).
				Return(nil, nil, assert.AnError).
				Times(1)
			suite.mockA2AMetrics.EXPECT().
				RecordSendMessage(ctx, sendEndpointUrl, gomock.Any(), assert.AnError, false).
				Times(1)
		}

		t.Run("Should return the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.send()

			assert.ErrorIs(t, err, assert.AnError)
		})
	})

	t.Run("Given the agent quotes a payment", func(t *testing.T) {
		t.Parallel()

		quote := &x402.PaymentRequired{X402Version: x402.Version}

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockA2AClient.EXPECT().
				SendMessage(ctx, sendEndpointUrl, gomock.Any(), sendSignature).
				Return(nil, quote, nil).
				Times(1)
			suite.mockA2AMetrics.EXPECT().
				RecordSendMessage(ctx, sendEndpointUrl, gomock.Any(), nil, true).
				Times(1)
		}

		t.Run("Should return the quote", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, paymentRequired, err := suite.send()

			assert.NoError(t, err)
			assert.Same(t, quote, paymentRequired)
		})
	})

	t.Run("Given the agent returns no reply", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockA2AClient.EXPECT().
				SendMessage(ctx, sendEndpointUrl, gomock.Any(), sendSignature).
				Return(nil, nil, nil).
				Times(1)
			suite.mockA2AMetrics.EXPECT().
				RecordSendMessage(ctx, sendEndpointUrl, gomock.Any(), nil, false).
				Times(1)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.send()

			assert.ErrorContains(t, err, "agent returned neither a reply nor a payment quote")
		})
	})

	t.Run("Given the agent answers with a json-rpc error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockA2AClient.EXPECT().
				SendMessage(ctx, sendEndpointUrl, gomock.Any(), sendSignature).
				Return(&models.ExternalA2AResponse{
					Error: &models.ExternalA2AError{Code: -32601, Message: "unsupported method"},
				}, nil, nil).
				Times(1)
			suite.mockA2AMetrics.EXPECT().
				RecordSendMessage(ctx, sendEndpointUrl, gomock.Any(), nil, false).
				Times(1)
		}

		t.Run("Should return its code and message", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.send()

			assert.ErrorContains(t, err, "agent returned error -32601: unsupported method")
		})
	})

	t.Run("Given the agent answers without a message", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockA2AClient.EXPECT().
				SendMessage(ctx, sendEndpointUrl, gomock.Any(), sendSignature).
				Return(&models.ExternalA2AResponse{Result: &models.ExternalSendMessageResult{}}, nil, nil).
				Times(1)
			suite.mockA2AMetrics.EXPECT().
				RecordSendMessage(ctx, sendEndpointUrl, gomock.Any(), nil, false).
				Times(1)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)

			_, _, err := suite.send()

			assert.ErrorContains(t, err, "agent replied without a message")
		})
	})

	t.Run("Given the agent answers", func(t *testing.T) {
		t.Parallel()

		response := &models.ExternalA2AResponse{
			Result: &models.ExternalSendMessageResult{
				Message: &models.ExternalA2AMessage{
					Parts:     []models.ExternalA2APart{{Text: "hi"}},
					ContextId: sendContextId,
				},
			},
		}

		initSuite := func(suite *WhenSendingA2AMessageTestingSuite) {
			suite.mockA2AMetrics.EXPECT().
				RecordSendMessage(ctx, sendEndpointUrl, gomock.Any(), nil, false).
				Times(1)
		}

		t.Run("Should send a SendMessage envelope from the user", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)
			suite.mockA2AClient.EXPECT().
				SendMessage(ctx, sendEndpointUrl, gomock.Any(), sendSignature).
				DoAndReturn(func(
					_ context.Context,
					_ string,
					request *models.ExternalA2ARequest,
					_ string,
				) (*models.ExternalA2AResponse, *x402.PaymentRequired, error) {
					assert.Equal(t, "SendMessage", request.Method)
					assert.Equal(t, "ROLE_USER", request.Params.Message.Role)
					assert.Equal(t, []models.ExternalA2APart{{Text: sendMessage}}, request.Params.Message.Parts)
					assert.Equal(t, sendContextId, request.Params.Message.ContextId)
					return response, nil, nil
				}).
				Times(1)

			_, _, err := suite.send()

			assert.NoError(t, err)
		})

		t.Run("Should return the reply", func(t *testing.T) {
			t.Parallel()

			suite := WhenSendingA2AMessageBeforeEach(t)
			initSuite(suite)
			suite.mockA2AClient.EXPECT().
				SendMessage(ctx, sendEndpointUrl, gomock.Any(), sendSignature).
				Return(response, nil, nil).
				Times(1)

			reply, _, err := suite.send()

			assert.NoError(t, err)
			assert.Equal(t, &interfaces.A2AReply{Text: "hi", ContextId: sendContextId}, reply)
		})
	})
}
