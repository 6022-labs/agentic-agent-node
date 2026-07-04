package services_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentGuidanceClientTestingSuite struct {
	sut *services.HttpAgentGuidanceClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenGettingAgentGuidanceClientBeforeEach(t *testing.T) *WhenGettingAgentGuidanceClientTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenGettingAgentGuidanceClientTestingSuite{
		sut:        services.NewHttpAgentGuidanceClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenGettingAgentGuidanceClient(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime returns the guidance", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceClientTestingSuite) {
			suite.mockClient.EXPECT().GetInternalAgentGuidanceWithResponse(gomock.Any()).Return(&clients.GetInternalAgentGuidanceResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusOK},
				Body:         []byte(`{"content":"be concise"}`),
			}, nil)
		}

		t.Run("Should decode the guidance payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceClientBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Get(context.Background())

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "be concise", payload.Content)
			}
		})
	})
}
