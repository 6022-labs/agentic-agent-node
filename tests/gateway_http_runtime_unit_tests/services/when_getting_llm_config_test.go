package services_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingLLMConfigTestingSuite struct {
	sut *services.HttpLLMConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingLLMConfigBeforeEach(t *testing.T) *WhenGettingLLMConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingLLMConfigTestingSuite{
		sut:        services.NewHttpLLMConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingLLMConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the configured connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigLlmWithResponse(gomock.Any()).Return(&clients.GetInternalConfigLlmResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"id":"llm-1","platform":"openai","model":"gpt-4o","setupStatus":"ready"}`),
			}, nil)
		}

		t.Run("Should decode the LLM connection payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConfigBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "llm-1", payload.Id)
				assert.Equal(t, "openai", payload.Platform)
				assert.Equal(t, "gpt-4o", payload.Model)
			}
		})
	})

	t.Run("Given no LLM connection is configured", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigLlmWithResponse(gomock.Any()).Return(&clients.GetInternalConfigLlmResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
				Body:         []byte(`{"code":"not_found"}`),
			}, nil)
		}

		t.Run("Should return ErrLLMConnectionNotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			assert.ErrorIs(t, err, gateway_interfaces.ErrLLMConnectionNotFound)
		})
	})

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().GetInternalConfigLlmWithResponse(gomock.Any()).Return(nil, errors.New("runtime down"))
		}

		t.Run("Should propagate the transport error", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingLLMConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			assert.ErrorContains(t, err, "runtime down")
		})
	})
}
