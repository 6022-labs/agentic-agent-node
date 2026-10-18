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

type WhenGettingAgentGuidanceTestingSuite struct {
	sut *services.HttpAgentGuidanceClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingAgentGuidanceBeforeEach(t *testing.T) *WhenGettingAgentGuidanceTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingAgentGuidanceTestingSuite{
		sut:        services.NewHttpAgentGuidanceClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingAgentGuidance(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentGuidanceWithResponse(gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentGuidanceWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentGuidanceResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Get(context.Background())

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime returns the guidance", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			suite.mockClient.EXPECT().
				GetInternalAgentGuidanceWithResponse(gomock.Any()).
				Return(&clients.GetInternalAgentGuidanceResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"content":"be concise"}`),
				}, nil)
		}

		t.Run("Should decode the guidance payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "be concise", payload.Content)
			}
		})
	})
}
