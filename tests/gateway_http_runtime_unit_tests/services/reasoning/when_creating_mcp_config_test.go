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

type WhenCreatingMCPConfigTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenCreatingMCPConfigBeforeEach(t *testing.T) *WhenCreatingMCPConfigTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenCreatingMCPConfigTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenCreatingMCPConfig(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalCreateMCPConnectionPayload{
				Name:       "tools",
				Url:        "https://mcp.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer x"},
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime answers 400", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConfigMcpResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalCreateMCPConnectionPayload{
				Name:       "tools",
				Url:        "https://mcp.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer x"},
			})

			var target *apperrors.ValidationError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 409", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConfigMcpResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusConflict},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 409 onto a ConflictError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalCreateMCPConnectionPayload{
				Name:       "tools",
				Url:        "https://mcp.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer x"},
			})

			var target *apperrors.ConflictError
			assert.ErrorAs(t, err, &target)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigMcpResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"boom"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConfigBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Create(context.Background(), &models.ExternalCreateMCPConnectionPayload{
				Name:       "tools",
				Url:        "https://mcp.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer x"},
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime accepts the new connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConfigTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(
					func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigMcpResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalCreateMCPConnectionPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Equal(t, "tools", sent.Name)
						assert.Equal(t, "Bearer x", sent.HeaderAuth["Authorization"])

						return &clients.PostInternalConfigMcpResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusOK},
							Body: []byte(
								`{"id":"mcp-1","name":"tools","url":"https://mcp.example","headerAuthKeys":["Authorization"]}`,
							),
						}, nil
					})
		}

		t.Run("Should decode the created MCP connection payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConfigBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.Create(context.Background(), &models.ExternalCreateMCPConnectionPayload{
				Name:       "tools",
				Url:        "https://mcp.example",
				HeaderAuth: map[string]string{"Authorization": "Bearer x"},
			})

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Equal(t, "mcp-1", payload.Id)
				assert.Equal(t, []string{"Authorization"}, payload.HeaderAuthKeys)
			}
		})
	})
}
