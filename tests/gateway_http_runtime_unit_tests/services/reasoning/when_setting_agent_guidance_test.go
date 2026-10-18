package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"

	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingAgentGuidanceTestingSuite struct {
	sut *services.HttpAgentGuidanceClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenSettingAgentGuidanceBeforeEach(t *testing.T) *WhenSettingAgentGuidanceTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenSettingAgentGuidanceTestingSuite{
		sut:        services.NewHttpAgentGuidanceClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenSettingAgentGuidance(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalAgentGuidanceWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Set(
				context.Background(),
				&models.ExternalAgentGuidancePayload{Content: "be concise"},
			)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalAgentGuidanceWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalAgentGuidanceResponse{HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Set(
				context.Background(),
				&models.ExternalAgentGuidancePayload{Content: "be concise"},
			)

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
			suite.mockClient.EXPECT().
				PutInternalAgentGuidanceWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PutInternalAgentGuidanceResponse{HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError}, Body: []byte(`{"code":"x","field":"f","message":"m"}`)}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Set(
				context.Background(),
				&models.ExternalAgentGuidancePayload{Content: "be concise"},
			)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime accepts the update", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
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

			suite := WhenSettingAgentGuidanceBeforeEach(t)
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
