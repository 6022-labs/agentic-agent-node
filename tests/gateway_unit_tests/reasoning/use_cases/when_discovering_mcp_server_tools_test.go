package use_cases_test

import (
	"context"
	"testing"

	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/gateway/reasoning/use_cases/requests"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/gateway_mocks/reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDiscoveringMcpServerToolsTestingSuite struct {
	sut *use_cases.DiscoverMcpServerTools

	mockClient *interfaces_mocks.MockMCPConfigClientInterface
}

func WhenDiscoveringMcpServerToolsBeforeEach(t *testing.T) *WhenDiscoveringMcpServerToolsTestingSuite {
	mockController := gomock.NewController(t)

	mockClient := interfaces_mocks.NewMockMCPConfigClientInterface(mockController)

	sut := use_cases.NewDiscoverMcpServerTools(mockClient)

	return &WhenDiscoveringMcpServerToolsTestingSuite{
		sut: sut,

		mockClient: mockClient,
	}
}

func TestWhenDiscoveringMcpServerTools(t *testing.T) {
	t.Parallel()

	t.Run("Given a request without a url", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without calling the runtime", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMcpServerToolsBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMcpServerToolsRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given a header with an empty name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError on headerAuth", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMcpServerToolsBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMcpServerToolsRequest{
				Url:        lo.ToPtr("https://mcp.github.example"),
				HeaderAuth: map[string]string{"  ": "Bearer abc"},
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "headerAuth", validationErr.Field)
		})
	})

	t.Run("Given a url with two auth headers", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMcpServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				DiscoverTools(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, body *models.ExternalDiscoverMCPServerToolsPayload) (*models.ExternalMCPServerToolsPayload, error) {
					assert.Equal(t, "https://mcp.github.example", lo.FromPtr(body.Url))
					assert.Equal(t, map[string]string{"Authorization": "Bearer abc", "X-Api-Key": "k"}, body.HeaderAuth)
					assert.Nil(t, body.ConnectionId)
					return &models.ExternalMCPServerToolsPayload{
						Tools: []models.ExternalMCPServerToolPayload{
							{Name: "search_issues", Description: "Search issues"},
						},
					}, nil
				})
		}

		t.Run("Should return the tools the runtime listed", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMcpServerToolsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.DiscoverMcpServerToolsRequest{
				Url:        lo.ToPtr("https://mcp.github.example"),
				HeaderAuth: map[string]string{"Authorization": "Bearer abc", "X-Api-Key": "k"},
			})

			assert.NoError(t, err)
			assert.Len(t, resp.Tools, 1)
			assert.Equal(t, "search_issues", resp.Tools[0].Name)
			assert.Equal(t, "Search issues", resp.Tools[0].Description)
		})
	})

	t.Run("Given a stored connection id and a header name with no value", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDiscoveringMcpServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				DiscoverTools(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, body *models.ExternalDiscoverMCPServerToolsPayload) (*models.ExternalMCPServerToolsPayload, error) {
					// The blank value travels: the runtime reads it as "use the stored value".
					assert.Equal(t, map[string]string{"Authorization": ""}, body.HeaderAuth)
					assert.Equal(t, id.String(), *body.ConnectionId)
					return &models.ExternalMCPServerToolsPayload{}, nil
				})
		}

		t.Run("Should forward the connection id and the valueless header", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMcpServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMcpServerToolsRequest{
				Url:          lo.ToPtr("https://mcp.github.example"),
				HeaderAuth:   map[string]string{"Authorization": ""},
				ConnectionId: &id,
			})

			assert.NoError(t, err)
		})
	})

	t.Run("Given no auth header at all", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDiscoveringMcpServerToolsTestingSuite) {
			suite.mockClient.EXPECT().
				DiscoverTools(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, body *models.ExternalDiscoverMCPServerToolsPayload) (*models.ExternalMCPServerToolsPayload, error) {
					assert.Nil(t, body.HeaderAuth)
					return &models.ExternalMCPServerToolsPayload{}, nil
				})
		}

		t.Run("Should pass nil so the runtime uses every stored header", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMcpServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.DiscoverMcpServerToolsRequest{ConnectionId: &id},
			)

			assert.NoError(t, err)
		})
	})
}
