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

type WhenCreatingResponseTestingSuite struct {
	sut *services.HttpResponseClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingResponseBeforeEach(t *testing.T) *WhenCreatingResponseTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingResponseTestingSuite{
		sut:        services.NewHttpResponseClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingResponse(t *testing.T) {
	t.Parallel()

	request := &models.ExternalCreateResponsePayload{
		PreviousResponseId: "conv-1",
		Input: []models.ExternalInputMessagePayload{
			{Type: "message", Role: "user", Content: "hello"},
		},
	}

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			var unavailableErr *apperrors.UnavailableError
			assert.ErrorAs(t, err, &unavailableErr)
			assert.Equal(t, "runtime_unavailable", unavailableErr.Code)
			assert.NotContains(t, unavailableErr.Message, "runtime down")
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalResponsesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"field":"input","message":"cannot be empty"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "input", validationErr.Field)
		})
	})

	t.Run("Given the runtime receives no caller id", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalResponsesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusUnauthorized},
					Body: []byte(
						`{"code":"missing_caller_id","message":"an authenticated caller id is required"}`,
					),
				}, nil)
		}

		t.Run("Should map the 401 onto an UnauthorizedError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			var unauthorizedErr *apperrors.UnauthorizedError
			assert.ErrorAs(t, err, &unauthorizedErr)
			assert.Equal(t, "missing_caller_id", unauthorizedErr.Code)
		})
	})

	t.Run("Given the conversation is out of the caller's view", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalResponsesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"conversation_unavailable","message":"taken"}`),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
			assert.Equal(t, "conversation_unavailable", conflictErr.Code)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalResponsesWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalResponsesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime returns a completed response", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) *models.ExternalCreateResponsePayload {
			sent := &models.ExternalCreateResponsePayload{}
			suite.mockClient.EXPECT().
				PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalResponsesResponse, error) {
					raw, _ := io.ReadAll(reader)
					_ = json.Unmarshal(raw, sent)
					return &clients.PostInternalResponsesResponse{
						HTTPResponse: &http.Response{StatusCode: http.StatusOK},
						Body: []byte(
							`{"id":"conv-1","object":"response","status":"completed",` +
								`"output":[{"type":"message","role":"assistant","content":[{"type":"output_text","text":"hi"}]}]}`,
						),
					}, nil
				})
			return sent
		}

		t.Run("Should send the request", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			sent := initSuite(suite)

			_, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, "conv-1", sent.PreviousResponseId)
		})

		t.Run("Should decode the OpenAI response", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "conv-1", result.Id)
				assert.Equal(t, "completed", result.Status)
				if assert.Len(t, result.Output, 1) && assert.Len(t, result.Output[0].Content, 1) {
					assert.Equal(t, "hi", result.Output[0].Content[0].Text)
				}
			}
		})
	})

	t.Run("Given the runtime returns an incomplete response", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingResponseTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalResponsesWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalResponsesResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusOK},
					Body:         []byte(`{"id":"conv-1","object":"response","status":"incomplete","output":[]}`),
				}, nil)
		}

		t.Run("Should return a result with no output", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingResponseBeforeEach(t)
			initSuite(suite)

			result, err := suite.sut.Create(context.Background(), request)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "incomplete", result.Status)
				assert.Empty(t, result.Output)
			}
		})
	})
}
