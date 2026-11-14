package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases/requests"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenUpdatingMCPConnectionTestingSuite struct {
	sut *use_cases.UpdateMCPConnection

	mockRepository    *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockMcpRequester  *interfaces_mocks.MockMcpRequesterInterface
	mockAgentResolver *identity_interfaces_mocks.MockAgentResolverInterface
}

func WhenUpdatingMCPConnectionBeforeEach(t *testing.T) *WhenUpdatingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockMcpRequester := interfaces_mocks.NewMockMcpRequesterInterface(mockController)
	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)

	sut := use_cases.NewUpdateMCPConnection(mockRepository, mockSecretStorage, mockMcpRequester, mockAgentResolver)

	return &WhenUpdatingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:    mockRepository,
		mockSecretStorage: mockSecretStorage,
		mockMcpRequester:  mockMcpRequester,
		mockAgentResolver: mockAgentResolver,
	}
}

func generateValidUpdateMCPConnectionRequest(id uuid.UUID) *requests.UpdateMCPConnectionRequest {
	return &requests.UpdateMCPConnectionRequest{
		Id: id,
		MCPConnectionRequest: requests.MCPConnectionRequest{
			Name:                "new-name",
			Url:                 "https://new",
			AllowedToolPatterns: []string{"search_issues"},
		},
	}
}

func TestWhenUpdatingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := generateValidUpdateMCPConnectionRequest(uuid.New())
		request.Name = ""

		t.Run("Should return a ValidationError without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the connection does not exist", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(nil, nil)
		}

		t.Run("Should return NotFound without touching update or vault", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateMCPConnectionRequest(id))

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "mcp_connection_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given another connection holds that name", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().WithId(id).WithName("old-name").Build()
			other := model_builders.NewMCPConnectionBuilder().WithName("taken").Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{*row, *other}, nil)
		}

		t.Run("Should return a conflict without updating", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMCPConnectionRequest(id)
			request.Name = "taken"

			_, err := suite.sut.Execute(context.Background(), request)

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})

	t.Run("Given a blank header value with nothing stored for it", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().WithId(id).WithName("new-name").Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().GetValue(gomock.Any(), id, "Authorization").Return("", false, nil)
		}

		t.Run("Should return a ValidationError without probing or updating", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMCPConnectionRequest(id)
			request.HeaderAuth = map[string]string{"Authorization": ""}

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "headerAuth", validationErr.Field)
		})
	})

	t.Run("Given the server cannot list its tools", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("new-name").
				WithHeaderAuthKeys([]string{}).
				Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("connection refused"))
		}

		t.Run(
			"Should return a validation error without touching the row or the vault",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenUpdatingMCPConnectionBeforeEach(t)
				initSuite(suite)

				_, err := suite.sut.Execute(context.Background(), generateValidUpdateMCPConnectionRequest(id))

				var validationErr *apperrors.ValidationError
				assert.ErrorAs(t, err, &validationErr)
				assert.Equal(t, "url", validationErr.Field)
				assert.Contains(t, validationErr.Message, "connection refused")
			},
		)
	})

	t.Run("Given a blank header value", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("new-name").
				WithHeaderAuthKeys([]string{"Authorization"}).
				Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSecretStorage.EXPECT().
				GetValue(gomock.Any(), id, "Authorization").
				Return("Bearer stored", true, nil)
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c *models.MCPConnection) ([]*models.Tool, error) {
					assert.Equal(t, map[string]string{"Authorization": "Bearer stored"}, c.HeaderAuth)
					return []*models.Tool{{Name: "search_issues"}}, nil
				})
			suite.mockRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c *models.MCPConnection) (*models.MCPConnection, error) {
					return c, nil
				})
			suite.mockSecretStorage.EXPECT().SaveValue(gomock.Any(), id, "Authorization", "Bearer stored").Return(nil)
		}

		t.Run("Should keep the value stored in vault", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMCPConnectionRequest(id)
			request.HeaderAuth = map[string]string{"Authorization": ""}

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the request drops every header", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("new-name").
				WithHeaderAuthKeys([]string{"Authorization", "X-Api-Key"}).
				Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c *models.MCPConnection) ([]*models.Tool, error) {
					assert.Empty(t, c.HeaderAuth)
					return []*models.Tool{{Name: "search_issues"}}, nil
				})
			suite.mockSecretStorage.EXPECT().DeleteValue(gomock.Any(), id, "Authorization").Return(nil)
			suite.mockSecretStorage.EXPECT().DeleteValue(gomock.Any(), id, "X-Api-Key").Return(nil)
			suite.mockRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c *models.MCPConnection) (*models.MCPConnection, error) {
					assert.Empty(t, c.HeaderAuthKeys)
					return c, nil
				})
		}

		t.Run("Should delete every stored secret", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMCPConnectionRequest(id)
			request.HeaderAuth = nil

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})

		t.Run("Should return a connection without header keys", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMCPConnectionRequest(id)
			request.HeaderAuth = nil

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Empty(t, resp.HeaderAuthKeys)
		})
	})

	t.Run("Given an existing connection with two headers", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().
				WithId(id).
				WithName("old-name").
				WithUrl("https://old").
				WithHeaderAuthKeys([]string{"Authorization", "X-Drop"}).
				Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return([]models.MCPConnection{*row}, nil)
			suite.mockMcpRequester.EXPECT().
				ListTools(gomock.Any(), gomock.Any()).
				Return([]*models.Tool{{Name: "search_issues"}}, nil)
			suite.mockSecretStorage.EXPECT().DeleteValue(gomock.Any(), id, "X-Drop").Return(nil)
			suite.mockRepository.EXPECT().
				Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c *models.MCPConnection) (*models.MCPConnection, error) {
					assert.Equal(t, "new-name", c.Name)
					assert.Equal(t, "https://new", c.Url)
					assert.Equal(t, []string{"search_issues"}, c.AllowedToolPatterns)
					assert.Len(t, c.HeaderAuthKeys, 1)
					assert.Equal(t, "Authorization", string(c.HeaderAuthKeys[0]))
					return c, nil
				})
			suite.mockSecretStorage.EXPECT().SaveValue(gomock.Any(), id, "Authorization", "Bearer new").Return(nil)
		}

		t.Run("Should update the row", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMCPConnectionRequest(id)
			request.HeaderAuth = map[string]string{"Authorization": "Bearer new"}

			_, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
		})

		t.Run("Should return the updated connection", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			request := generateValidUpdateMCPConnectionRequest(id)
			request.HeaderAuth = map[string]string{"Authorization": "Bearer new"}

			resp, err := suite.sut.Execute(context.Background(), request)

			assert.NoError(t, err)
			assert.Equal(t, id, resp.Id)
		})
	})
}
