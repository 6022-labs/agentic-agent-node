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

type WhenUpdatingSelfToolTestingSuite struct {
	sut *use_cases.UpdateSelfTool

	mockSelfToolRepository *repositories_mocks.MockSelfToolRepositoryInterface
	mockSelfToolRegistry   *interfaces_mocks.MockSelfToolRegistryInterface
	mockAgentResolver      *identity_interfaces_mocks.MockAgentResolverInterface
}

func WhenUpdatingSelfToolBeforeEach(t *testing.T) *WhenUpdatingSelfToolTestingSuite {
	mockController := gomock.NewController(t)

	mockSelfToolRepository := repositories_mocks.NewMockSelfToolRepositoryInterface(mockController)
	mockSelfToolRegistry := interfaces_mocks.NewMockSelfToolRegistryInterface(mockController)
	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)

	sut := use_cases.NewUpdateSelfTool(mockSelfToolRepository, mockSelfToolRegistry, mockAgentResolver)

	return &WhenUpdatingSelfToolTestingSuite{
		sut: sut,

		mockSelfToolRepository: mockSelfToolRepository,
		mockSelfToolRegistry:   mockSelfToolRegistry,
		mockAgentResolver:      mockAgentResolver,
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

	t.Run("Given an invalid request", func(t *testing.T) {
		t.Parallel()

		request := generateValidUpdateSelfToolRequest(uuid.New(), false)
		request.Enabled = nil

		t.Run("Should return a ValidationError without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)

			_, err := suite.sut.Execute(context.Background(), request)

			var validationErr *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationErr)
		})
	})

	t.Run("Given the self tool lookup fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().
				GetById(gomock.Any(), gomock.Any(), id).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateSelfToolRequest(id, true))

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "get_failed", internalErr.Code)
		})
	})

	t.Run("Given the self tool does not exist", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
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

	t.Run("Given saving the self tool fails", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()
		initSuite := func(suite *WhenUpdatingSelfToolTestingSuite) {
			row := model_builders.NewSelfToolBuilder().WithId(id).WithEnabled(true).Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().GetById(gomock.Any(), gomock.Any(), id).Return(row, nil)
			suite.mockSelfToolRepository.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError without reading the self tool definitions", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateSelfToolRequest(id, false))

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "update_failed", internalErr.Code)
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
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
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

		t.Run("Should toggle it off", func(t *testing.T) {
			t.Parallel()

			suite := WhenUpdatingSelfToolBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), generateValidUpdateSelfToolRequest(id, false))

			assert.NoError(t, err)
		})

		t.Run("Should return the joined response", func(t *testing.T) {
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
}
