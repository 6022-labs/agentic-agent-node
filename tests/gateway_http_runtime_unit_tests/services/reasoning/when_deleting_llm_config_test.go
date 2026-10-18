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

type WhenDeletingLLMConfigTestingSuite struct {
	sut *services.HttpLLMConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDeletingLLMConfigBeforeEach(t *testing.T) *WhenDeletingLLMConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDeletingLLMConfigTestingSuite{
		sut:        services.NewHttpLLMConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDeletingLLMConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigLlmWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigLlmWithResponse(gomock.Any()).
				Return(&clients.DeleteInternalConfigLlmResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime deletes the connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDeletingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().
				DeleteInternalConfigLlmWithResponse(gomock.Any()).
				Return(&clients.DeleteInternalConfigLlmResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
				}, nil)
		}

		t.Run("Should delete the LLM connection without error", func(t *testing.T) {
			t.Parallel()

			suite := WhenDeletingLLMConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Delete(context.Background())

			assert.NoError(t, err)
		})
	})
}
