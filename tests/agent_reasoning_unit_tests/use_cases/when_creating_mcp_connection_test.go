package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenCreatingMCPConnectionTestingSuite struct {
	sut *use_cases.CreateMCPConnection

	mockRepository      *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage   *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockMcpRequester    *interfaces_mocks.MockMcpRequesterInterface
	mockAgentRepository *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenCreatingMCPConnectionBeforeEach(t *testing.T) *WhenCreatingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockMcpRequester := interfaces_mocks.NewMockMcpRequesterInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewCreateMCPConnection(mockRepository, mockSecretStorage, mockMcpRequester, mockAgentRepository)

	return &WhenCreatingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:      mockRepository,
		mockSecretStorage:   mockSecretStorage,
		mockMcpRequester:    mockMcpRequester,
		mockAgentRepository: mockAgentRepository,
	}
}

func generateValidCreateMCPConnectionRequest() *requests.CreateMCPConnectionRequest {
	return &requests.CreateMCPConnectionRequest{
		Name:                "github",
		Url:                 "https://mcp.github.example",
		AllowedToolPatterns: []string{"search_issues"},
	}
}

func TestWhenCreatingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)

			request := generateValidCreateMCPConnectionRequest()
			request.Name = ""

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given no allowed tool pattern", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)

			request := generateValidCreateMCPConnectionRequest()
			request.AllowedToolPatterns = nil

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "allowedToolPatterns", validationErr.Field)
		})
	})

	t.Run("Given a request with one header", func(t *testing.T) {
		t.Parallel()

		generatedId := uuid.New()
		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) ([]*models.Tool, error) {
				assert.Equal(t, "https://mcp.github.example", c.Url)
				assert.Equal(t, map[string]string{"Authorization": "Bearer abc"}, c.HeaderAuth)
				return []*models.Tool{{Name: "search_issues"}}, nil
			})
			suite.mockRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) (*models.MCPConnection, error) {
				assert.Equal(t, "github", c.Name)
				assert.Equal(t, "https://mcp.github.example", c.Url)
				assert.Equal(t, []string{"search_issues"}, c.AllowedToolPatterns)
				assert.Len(t, c.HeaderAuthKeys, 1)
				assert.Equal(t, "Authorization", string(c.HeaderAuthKeys[0]))
				c.Id = generatedId
				return c, nil
			})
			suite.mockSecretStorage.EXPECT().SaveValue(gomock.Any(), generatedId, "Authorization", "Bearer abc").Return(nil)
		}

		t.Run("Should probe the server, persist the row and the header value under the generated id", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			request := generateValidCreateMCPConnectionRequest()
			request.HeaderAuth = map[string]string{"Authorization": "Bearer abc"}
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, generatedId, resp.Id)
			assert.ElementsMatch(t, []string{"Authorization"}, resp.HeaderAuthKeys)
			assert.Equal(t, []string{"search_issues"}, resp.AllowedToolPatterns)
		})
	})

	t.Run("Given a request with no headers", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).Return([]*models.Tool{{Name: "search_issues"}}, nil)
			suite.mockRepository.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) (*models.MCPConnection, error) {
				assert.Empty(t, c.HeaderAuthKeys)
				c.Id = uuid.New()
				return c, nil
			})
		}

		t.Run("Should not call SaveValue and return Created", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			request := generateValidCreateMCPConnectionRequest()
			request.Name = "anon"
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	})

	t.Run("Given the server cannot list its tools", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).Return(nil, errors.New("connection refused"))
		}

		t.Run("Should reject the save with a validation error before creating the row", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidCreateMCPConnectionRequest())

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "url", validationErr.Field)
			assert.Contains(t, validationErr.Message, "connection refused")
		})
	})

	t.Run("Given the name collides with an existing connection", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			existing := model_builders.NewMCPConnectionBuilder().WithName("github").Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return([]models.MCPConnection{*existing}, nil)
		}

		t.Run("Should return a conflict without probing or creating", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidCreateMCPConnectionRequest())

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})

	t.Run("Given the repository create fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCreatingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return([]models.MCPConnection{}, nil)
			suite.mockMcpRequester.EXPECT().ListTools(gomock.Any(), gomock.Any()).Return([]*models.Tool{{Name: "search_issues"}}, nil)
			suite.mockRepository.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should not touch the secret storage and return InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenCreatingMCPConnectionBeforeEach(t)
			request := generateValidCreateMCPConnectionRequest()
			request.HeaderAuth = map[string]string{"K": "V"}
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), request)

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db down")
		})
	})
}
