package use_cases_test

import (
	"context"
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

type WhenUpdatingMCPConnectionTestingSuite struct {
	sut *use_cases.UpdateMCPConnection

	mockRepository      *repositories_mocks.MockMCPConnectionRepositoryInterface
	mockSecretStorage   *interfaces_mocks.MockMCPHeaderAuthSecretStorageInterface
	mockAgentRepository *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenUpdatingMCPConnectionBeforeEach(t *testing.T) *WhenUpdatingMCPConnectionTestingSuite {
	mockController := gomock.NewController(t)

	mockRepository := repositories_mocks.NewMockMCPConnectionRepositoryInterface(mockController)
	mockSecretStorage := interfaces_mocks.NewMockMCPHeaderAuthSecretStorageInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewUpdateMCPConnection(mockRepository, mockSecretStorage, mockAgentRepository)

	return &WhenUpdatingMCPConnectionTestingSuite{
		sut: sut,

		mockRepository:      mockRepository,
		mockSecretStorage:   mockSecretStorage,
		mockAgentRepository: mockAgentRepository,
	}
}

func TestWhenUpdatingMCPConnection(t *testing.T) {
	t.Parallel()

	t.Run("Given the request is invalid", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), uuid.New(), &requests.SaveMCPConnectionRequest{Url: "https://my.mcp.com/v1"})

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given an unknown id", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(nil, nil)
		}

		t.Run("Should return NotFound without touching update or vault", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), id, &requests.SaveMCPConnectionRequest{Name: "x", Url: "https://x"})

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "mcp_connection_not_found", notFoundErr.Code)
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
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return([]models.MCPConnection{*row}, nil)
			suite.mockSecretStorage.EXPECT().DeleteValue(gomock.Any(), id, "X-Drop").Return(nil)
			suite.mockRepository.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, c *models.MCPConnection) (*models.MCPConnection, error) {
				assert.Equal(t, "new-name", c.Name)
				assert.Equal(t, "https://new", c.Url)
				assert.Len(t, c.HeaderAuthKeys, 1)
				assert.Equal(t, "Authorization", string(c.HeaderAuthKeys[0]))
				return c, nil
			})
			suite.mockSecretStorage.EXPECT().SaveValue(gomock.Any(), id, "Authorization", "Bearer new").Return(nil)
		}

		t.Run("Should drop removed header secrets, update the row, and return Ok", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), id, &requests.SaveMCPConnectionRequest{
				Name:       "new-name",
				Url:        "https://new",
				HeaderAuth: map[string]string{"Authorization": "Bearer new"},
			})

			assert.NoError(t, err)
			assert.Equal(t, id, resp.Id)
		})
	})

	t.Run("Given the new name collides with another connection", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingMCPConnectionTestingSuite) {
			row := model_builders.NewMCPConnectionBuilder().WithId(id).WithName("old-name").Build()
			other := model_builders.NewMCPConnectionBuilder().WithName("taken").Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return([]models.MCPConnection{*row, *other}, nil)
		}

		t.Run("Should return a conflict without updating", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingMCPConnectionBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), id, &requests.SaveMCPConnectionRequest{Name: "taken", Url: "https://x"})

			var conflictErr *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictErr)
		})
	})
}
