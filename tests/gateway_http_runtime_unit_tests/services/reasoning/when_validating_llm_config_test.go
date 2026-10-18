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

type WhenValidatingLLMConfigTestingSuite struct {
	sut *services.HttpLLMConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenValidatingLLMConfigBeforeEach(t *testing.T) *WhenValidatingLLMConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenValidatingLLMConfigTestingSuite{
		sut:        services.NewHttpLLMConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenValidatingLLMConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigLlmValidateWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Validate(context.Background(), &models.ExternalSetupLLMConnectionPayload{
				Platform: "openai",
				Model:    "gpt-4o",
				ApiKey:   "secret",
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigLlmValidateWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigLlmValidateResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body: []byte(
						`{"code":"validation_error","field":"connection","message":"401 unauthorized"}`,
					),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Validate(context.Background(), &models.ExternalSetupLLMConnectionPayload{
				Platform: "openai",
				Model:    "gpt-4o",
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "connection", validationErr.Field)
		})
	})

	t.Run("Given the runtime answers 502", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigLlmValidateWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigLlmValidateResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadGateway},
					Body:         []byte(`{"code":"boom","message":"runtime down"}`),
				}, nil)
		}

		t.Run("Should return an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Validate(
				context.Background(),
				&models.ExternalSetupLLMConnectionPayload{Platform: "openai", Model: "gpt-4o"},
			)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
		})
	})

	t.Run("Given the runtime accepts the connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenValidatingLLMConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigLlmValidateWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(
					func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigLlmValidateResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalSetupLLMConnectionPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Equal(t, "openai", sent.Platform)
						assert.Equal(t, "secret", sent.ApiKey)

						return &clients.PostInternalConfigLlmValidateResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusNoContent},
						}, nil
					})
		}

		t.Run("Should return no error", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingLLMConfigBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Validate(context.Background(), &models.ExternalSetupLLMConnectionPayload{
				Platform: "openai",
				Model:    "gpt-4o",
				ApiKey:   "secret",
			})

			assert.NoError(t, err)
		})
	})
}
