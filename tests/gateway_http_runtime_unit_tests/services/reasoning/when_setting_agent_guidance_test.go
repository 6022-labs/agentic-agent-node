package services_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingAgentGuidanceClientTestingSuite struct {
	sut *services.HttpAgentGuidanceClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenSettingAgentGuidanceClientBeforeEach(t *testing.T) *WhenSettingAgentGuidanceClientTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenSettingAgentGuidanceClientTestingSuite{
		sut:        services.NewHttpAgentGuidanceClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenSettingAgentGuidanceClient(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the update", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceClientTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalAgentGuidanceWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, contentType string, body io.Reader, _ ...clients.RequestEditorFn) (*clients.PutInternalAgentGuidanceResponse, error) {
					assert.Contains(t, contentType, "application/json")
					var sent models.ExternalAgentGuidancePayload
					assert.NoError(t, json.NewDecoder(body).Decode(&sent))
					assert.Equal(t, "be concise", sent.Content)
					return &clients.PutInternalAgentGuidanceResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"content":"be concise"}`),
					}, nil
				})
		}

		t.Run("Should send the JSON body and decode the echoed content", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceClientBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Set(
				context.Background(),
				&models.ExternalAgentGuidancePayload{Content: "be concise"},
			)

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "be concise", payload.Content)
			}
		})
	})
}
