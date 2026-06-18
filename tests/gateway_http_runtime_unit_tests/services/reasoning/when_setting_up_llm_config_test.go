package services_test

import (
	"context"
	"encoding/json"
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

type WhenSettingUpLLMConfigTestingSuite struct {
	sut *services.HttpLLMConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenSettingUpLLMConfigBeforeEach(t *testing.T) *WhenSettingUpLLMConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenSettingUpLLMConfigTestingSuite{
		sut:        services.NewHttpLLMConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenSettingUpLLMConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime accepts the setup", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().PostInternalConfigLlmWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigLlmResponse, error) {
					raw, err := io.ReadAll(reader)
					assert.NoError(t, err)

					var sent models.ExternalSetupLLMConnectionPayload
					assert.NoError(t, json.Unmarshal(raw, &sent))
					assert.Equal(t, "openai", sent.Platform)
					assert.Equal(t, "secret", sent.ApiKey)

					return &clients.PostInternalConfigLlmResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body:         []byte(`{"id":"llm-1","platform":"openai","model":"gpt-4o","setupStatus":"ready"}`),
					}, nil
				})
		}

		t.Run("Should decode the created LLM connection payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConfigBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Setup(context.Background(), &models.ExternalSetupLLMConnectionPayload{
				Platform: "openai",
				Model:    "gpt-4o",
				ApiKey:   "secret",
			})

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "llm-1", payload.Id)
			}
		})
	})

	t.Run("Given the runtime rejects the setup", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().PostInternalConfigLlmWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).Return(&clients.PostInternalConfigLlmResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
				Body:         []byte(`{"code":"boom","message":"runtime down"}`),
			}, nil)
		}

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Setup(context.Background(), &models.ExternalSetupLLMConnectionPayload{
				Platform: "openai",
				Model:    "gpt-4o",
				ApiKey:   "secret",
			})

			assert.Error(t, err)
		})
	})

	t.Run("Given the runtime rejects the setup with a validation error", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingUpLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().PostInternalConfigLlmWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).Return(&clients.PostInternalConfigLlmResponse{
				HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
				Body:         []byte(`{"code":"validation_error","field":"apiKey","message":"cannot be empty on first setup"}`),
			}, nil)
		}

		t.Run("Should map it to a ValidationError preserving the field", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingUpLLMConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Setup(context.Background(), &models.ExternalSetupLLMConnectionPayload{
				Platform: "openai",
				Model:    "gpt-4o",
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})
}
