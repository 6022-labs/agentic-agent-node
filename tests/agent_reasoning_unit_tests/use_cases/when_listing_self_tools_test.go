package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenListingSelfToolsTestingSuite struct {
	sut *use_cases.ListSelfTools

	mockSelfToolRepository *repositories_mocks.MockSelfToolRepositoryInterface
	mockSelfToolRegistry   *interfaces_mocks.MockSelfToolRegistryInterface
	mockAgentRepository    *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenListingSelfToolsBeforeEach(t *testing.T) *WhenListingSelfToolsTestingSuite {
	mockController := gomock.NewController(t)

	mockSelfToolRepository := repositories_mocks.NewMockSelfToolRepositoryInterface(mockController)
	mockSelfToolRegistry := interfaces_mocks.NewMockSelfToolRegistryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := use_cases.NewListSelfTools(mockSelfToolRepository, mockSelfToolRegistry, mockAgentRepository)

	return &WhenListingSelfToolsTestingSuite{
		sut: sut,

		mockSelfToolRepository: mockSelfToolRepository,
		mockSelfToolRegistry:   mockSelfToolRegistry,
		mockAgentRepository:    mockAgentRepository,
	}
}

func TestWhenListingSelfTools(t *testing.T) {
	t.Parallel()

	t.Run("Given listing the self tool rows fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenListingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().
				ListByAgentId(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "list_failed", internalErr.Code)
		})
	})

	t.Run("Given every row has a matching self tool definition", func(t *testing.T) {
		t.Parallel()

		rows := []models.SelfTool{
			*model_builders.NewSelfToolBuilder().WithName("get_identity").WithEnabled(true).Build(),
			*model_builders.NewSelfToolBuilder().WithName("run_command").WithEnabled(false).Build(),
		}
		initSuite := func(suite *WhenListingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(model_builders.NewAgentBuilder().Build(), nil)
			suite.mockSelfToolRepository.EXPECT().ListByAgentId(gomock.Any(), gomock.Any()).Return(rows, nil)
			suite.mockSelfToolRegistry.EXPECT().ListTools().Return([]*models.Tool{
				{Name: "get_identity", Description: "read own identity"},
				{Name: "run_command", Description: "run a command"},
			})
			suite.mockSelfToolRegistry.EXPECT().ListGroupNames().Return(map[string]string{
				"get_identity": "Identity",
				"run_command":  "Machine",
			})
		}

		t.Run("Should return the rows joined with their descriptions", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			if assert.Len(t, resp.Tools, 2) {
				assert.Equal(t, rows[0].Id, resp.Tools[0].Id)
				assert.Equal(t, "get_identity", resp.Tools[0].Name)
				assert.Equal(t, "read own identity", resp.Tools[0].Description)
				assert.Equal(t, "Identity", resp.Tools[0].Group)
				assert.True(t, resp.Tools[0].Enabled)

				assert.Equal(t, "run_command", resp.Tools[1].Name)
				assert.Equal(t, "run a command", resp.Tools[1].Description)
				assert.Equal(t, "Machine", resp.Tools[1].Group)
				assert.False(t, resp.Tools[1].Enabled)
			}
		})
	})
}
