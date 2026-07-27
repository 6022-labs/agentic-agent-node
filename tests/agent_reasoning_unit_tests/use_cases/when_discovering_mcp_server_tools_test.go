package use_cases_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	reasoning_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/responses"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenDiscoveringMCPServerToolsTestingSuite struct {
	sut *use_cases.DiscoverMCPServerTools

	mockRepository      *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage   *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockMcpRequester    *interfaces_mocks.MockMcpRequesterInterface
	mockAgentRepository *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenDiscoveringMCPServerToolsBeforeEach(t *testing.T) *WhenDiscoveringMCPServerToolsTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockMcpRequester := interfaces_mocks.NewMockMcpRequesterInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewDiscoverMCPServerTools(mockRepository, mockSecretStorage, mockMcpRequester, mockAgentRepository)

	return &WhenDiscoveringMCPServerToolsTestingSuite{
		sut: sut,

		mockRepository:      mockRepository,
		mockSecretStorage:   mockSecretStorage,
		mockMcpRequester:    mockMcpRequester,
		mockAgentRepository: mockAgentRepository,
	}
}

func TestWhenDiscoveringMCPServerTools(t *testing.T) {
	t.Parallel()

	t.Run("Given the url is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "url", validationErr.Field)
		})
	})

	t.Run("Given a url and header values", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) ([]*models.Tool, error) {
				assert.Equal(t, "https://mcp.github.example", c.Url)
				assert.Equal(t, map[string]string{"Authorization": "Bearer abc"}, c.HeaderAuth)
				return []*models.Tool{
					{Name: "search_issues", Description: "Search issues"},
					{Name: "create_issue"},
				}, nil
			})
		}

		t.Run("Should probe with the request headers and return the tool names", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{
				Url:        lo.ToPtr("https://mcp.github.example"),
				HeaderAuth: map[string]string{"Authorization": "Bearer abc"},
			})

			assert.NoError(t, err)
			assert.Equal(t, []string{"search_issues", "create_issue"}, lo.Map(resp.Tools, func(tool responses.MCPServerToolResponse, _ int) string { return tool.Name }))
			assert.Equal(t, "Search issues", resp.Tools[0].Description)
		})
	})

	t.Run("Given a stored connection id and no header value", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("github").
				WithHeaderAuthKeys([]string{"Authorization"}).
				Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "Authorization").Return("Bearer stored", true, nil)
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) ([]*models.Tool, error) {
				assert.Equal(t, "github", c.Name)
				assert.Equal(t, map[string]string{"Authorization": "Bearer stored"}, c.HeaderAuth)
				return []*models.Tool{{Name: "search_issues"}}, nil
			})
		}

		t.Run("Should hydrate the headers from vault", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{
				Url:          lo.ToPtr("https://mcp.github.example"),
				ConnectionId: &id,
			})

			assert.NoError(t, err)
			assert.Len(t, resp.Tools, 1)
		})
	})

	t.Run("Given an unknown connection id", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(nil, nil)
		}

		t.Run("Should return NotFound without probing", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{
				Url:          lo.ToPtr("https://mcp.github.example"),
				ConnectionId: &id,
			})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})

	t.Run("Given the server cannot list its tools", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection refused"))
		}

		t.Run("Should return a validation error carrying the reason", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{Url: lo.ToPtr("https://mcp.github.example")})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Contains(t, validationErr.Message, "connection refused")
		})
	})

	t.Run("Given the server refuses the credentials", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).
				Return(nil, fmt.Errorf("failed to list MCP tools: %w", reasoning_interfaces.ErrMCPUnauthorized))
		}

		t.Run("Should blame the headers instead of leaking the status code", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{
				Url:        lo.ToPtr("https://mcp.github.example"),
				HeaderAuth: map[string]string{"Authorization": "Bearer wrong"},
			})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "headerAuth", validationErr.Field)
			assert.Equal(t, "the server rejected these credentials — check the API key", validationErr.Message)
		})
	})

	t.Run("Given a stored connection id and no url", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("github").
				WithUrl("https://stored.mcp.example").
				WithHeaderAuthKeys([]string{}).
				Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) ([]*models.Tool, error) {
				assert.Equal(t, "https://stored.mcp.example", c.Url)
				return []*models.Tool{{Name: "search_issues"}}, nil
			})
		}

		t.Run("Should probe the stored url", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{ConnectionId: &id})

			assert.NoError(t, err)
		})
	})

	t.Run("Given a stored connection id and a changed url", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenDiscoveringMCPServerToolsTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("github").
				WithUrl("https://stored.mcp.example").
				WithHeaderAuthKeys([]string{"Authorization"}).
				Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "Authorization").Return("Bearer stored", true, nil)
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) ([]*models.Tool, error) {
				assert.Equal(t, "https://new.mcp.example", c.Url)
				assert.Equal(t, map[string]string{"Authorization": "Bearer stored"}, c.HeaderAuth)
				return []*models.Tool{{Name: "search_issues"}}, nil
			})
		}

		t.Run("Should probe the new url with the stored header value", func(t *testing.T) {
			t.Parallel()

			suite := WhenDiscoveringMCPServerToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.DiscoverMCPServerToolsRequest{
				ConnectionId: &id,
				Url:          lo.ToPtr("https://new.mcp.example"),
			})

			assert.NoError(t, err)
		})
	})
}
