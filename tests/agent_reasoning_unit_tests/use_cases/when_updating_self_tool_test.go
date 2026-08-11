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

type WhenUpdatingSelfToolTestingSuite struct {
	sut *use_cases.UpdateSelfTool

	mockSelfToolRepository *repositories_mocks.MockSelfToolRepositoryInterface
	mockSelfToolRegistry   *interfaces_mocks.MockSelfToolRegistryInterface
	mockAgentRepository    *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenUpdatingSelfToolBeforeEach(t *testing.T) *WhenUpdatingSelfToolTestingSuite {
	mockController := gomock.NewController(t)

	mockSelfToolRepository := repositories_mocks.NewMockSelfToolRepositoryInterface(mockController)
	mockSelfToolRegistry := interfaces_mocks.NewMockSelfToolRegistryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewUpdateSelfTool(mockSelfToolRepository, mockSelfToolRegistry, mockAgentRepository)

	return &WhenUpdatingSelfToolTestingSuite{
		sut: sut,

		mockSelfToolRepository: mockSelfToolRepository,
		mockSelfToolRegistry:   mockSelfToolRegistry,
		mockAgentRepository:    mockAgentRepository,
	}
}

func generateValidUpdateSelfToolRequest(id uuid.UUID, enabled bool) *requests.UpdateSelfToolRequest {
	return &requests.UpdateSelfToolRequest{
		Id:      id,
		Enabled: &enabled,
	}
}

func TestWhenUpdatingSelfTool(t *testing.T) {
	t.Parallel()

	t.Run("Given the enabled flag is missing", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a validation error without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)

			request := generateValidUpdateSelfToolRequest(uuid.New(), false)
			request.Enabled = nil

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
			assert.Equal(t, "enabled", validationErr.Field)
		})
	})

	t.Run("Given an unknown id", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(nil, nil)
		}

		t.Run("Should return NotFound without updating", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateSelfToolRequest(id, false))

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "self_tool_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given an existing enabled tool", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			row := model_builders.NewSelfToolBuilder().
				WithId(id).
				WithName("run_command").
				WithEnabled(true).
				Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSelfToolRepository.EXPECT().Update(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, tool *models.SelfTool) (*models.SelfTool, error) {
					assert.Equal(t, id, tool.Id)
					assert.False(t, tool.Enabled)
					return tool, nil
				})
			suite.mockSelfToolRegistry.EXPECT().ListTools().Return([]*models.Tool{
				{Name: "run_command", Description: "run a command"},
			})
			suite.mockSelfToolRegistry.EXPECT().ListGroupNames().Return(map[string]string{"run_command": "Machine"})
		}

		t.Run("Should toggle it off and return the joined response", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background(), generateValidUpdateSelfToolRequest(id, false))

			assert.NoError(t, err)
			assert.Equal(t, id, resp.Id)
			assert.Equal(t, "run_command", resp.Name)
			assert.Equal(t, "run a command", resp.Description)
			assert.Equal(t, "Machine", resp.Group)
			assert.False(t, resp.Enabled)
		})
	})

	t.Run("Given the row lookup fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any(), id).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateSelfToolRequest(id, true))

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "get_failed", internalErr.Code)
		})
	})

	t.Run("Given the update fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			row := model_builders.NewSelfToolBuilder().WithId(id).WithEnabled(true).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSelfToolRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an internal error without reading the self tool definitions", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateSelfToolRequest(id, false))

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
		})
	})
}
