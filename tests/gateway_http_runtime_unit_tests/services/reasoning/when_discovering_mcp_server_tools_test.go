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
	gateway_interfaces "github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/clients"
	services "github.com/6022-labs/agentic-agent-node/src/gateway_http_runtime/services/reasoning"
	clients_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_http_runtime_mocks/clients_mocks"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDiscoveringMCPServerToolsTestingSuite struct {
	sut *services.HttpMCPConfigClient

	mockClient *clients_mocks.MockClientWithResponsesInterface
}

func WhenDiscoveringMCPServerToolsBeforeEach(t *testing.T) *WhenDiscoveringMCPServerToolsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := clients_mocks.NewMockClientWithResponsesInterface(mockController)

	return &WhenDiscoveringMCPServerToolsTestingSuite{
		sut:        services.NewHttpMCPConfigClient(mockClient),
		mockClient: mockClient,
	}
}

func TestWhenDiscoveringMCPServerTools(t *testing.T) {
	t.Parallel()

	t.Run("Given the runtime call fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpToolsWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, errors.New("runtime down"))
		}

		t.Run("Should map the transport error onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.DiscoverTools(context.Background(), &models.ExternalDiscoverMCPServerToolsPayload{
				Url: lo.ToPtr("https://mcp.github.example"),
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the connection does not exist", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpToolsWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigMcpToolsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusNotFound},
					Body:         []byte(`{"code":"mcp_connection_not_found"}`),
				}, nil)
		}

		t.Run("Should return the not-found sentinel untouched", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.DiscoverTools(
				context.Background(),
				&models.ExternalDiscoverMCPServerToolsPayload{Url: lo.ToPtr("https://mcp.github.example")},
			)

			assert.ErrorIs(t, err, gateway_interfaces.ErrMCPConnectionNotFound)
		})
	})

	t.Run("Given the server cannot be reached", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpToolsWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				Return(&clients.PostInternalConfigMcpToolsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusBadRequest},
					Body: []byte(
						`{"code":"validation_error","field":"url","message":"failed to list tools: connection refused"}`,
					),
				}, nil)
		}

		t.Run("Should map the 400 onto a ValidationError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.DiscoverTools(
				context.Background(),
				&models.ExternalDiscoverMCPServerToolsPayload{Url: lo.ToPtr("https://mcp.github.example")},
			)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "url", validationErr.Field)
		})
	})

	t.Run("Given the runtime answers 500", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpToolsWithBodyWithResponse(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&clients.PostInternalConfigMcpToolsResponse{
					HTTPResponse: &http.Response{StatusCode: http.StatusInternalServerError},
					Body:         []byte(`{"code":"x","field":"f","message":"m"}`),
				}, nil)
		}

		t.Run("Should map the 500 onto an UnavailableError", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.DiscoverTools(context.Background(), &models.ExternalDiscoverMCPServerToolsPayload{
				Url: lo.ToPtr("https://mcp.github.example"),
			})

			var target *apperrors.UnavailableError
			if assert.ErrorAs(t, err, &target) {
				assert.Equal(t, "runtime_unavailable", target.Code)
			}
		})
	})

	t.Run("Given the runtime lists the tools", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				PostInternalConfigMcpToolsWithBodyWithResponse(gomock.Any(), "application/json", gomock.Any()).
				DoAndReturn(
					func(_ context.Context, _ string, reader io.Reader, _ ...clients.RequestEditorFn) (*clients.PostInternalConfigMcpToolsResponse, error) {
						raw, err := io.ReadAll(reader)
						assert.NoError(t, err)

						var sent models.ExternalDiscoverMCPServerToolsPayload
						assert.NoError(t, json.Unmarshal(raw, &sent))
						assert.Equal(t, "https://mcp.github.example", lo.FromPtr(sent.Url))

						return &clients.PostInternalConfigMcpToolsResponse{
							HTTPResponse: &http.Response{StatusCode: http.StatusOK},
							Body:         []byte(`{"tools":[{"name":"search_issues","description":"Search issues"}]}`),
						}, nil
					})
		}

		t.Run("Should decode the tools payload", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			payload, err := suite.sut.DiscoverTools(context.Background(), &models.ExternalDiscoverMCPServerToolsPayload{
				Url: lo.ToPtr("https://mcp.github.example"),
			})

			assert.NoError(t, err)
			if assert.NotNil(t, payload) {
				assert.Len(t, payload.Tools, 1)
				assert.Equal(t, "search_issues", payload.Tools[0].Name)
			}
		})
	})
}
