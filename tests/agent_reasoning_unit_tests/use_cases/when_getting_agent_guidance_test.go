package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	identity_repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentGuidanceTestingSuite struct {
	sut *use_cases.GetAgentGuidance

	mockAgentGuidanceRepository *repositories_mocks.MockAgentGuidanceRepositoryInterface
	mockAgentRepository         *identity_repository_mocks.MockAgentRepositoryInterface
}

func WhenGettingAgentGuidanceBeforeEach(t *testing.T) *WhenGettingAgentGuidanceTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentGuidanceRepository := repositories_mocks.NewMockAgentGuidanceRepositoryInterface(mockController)
	mockAgentRepository := identity_repository_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenGettingAgentGuidanceTestingSuite{
		sut:                         use_cases.NewGetAgentGuidance(mockAgentGuidanceRepository, mockAgentRepository),
		mockAgentGuidanceRepository: mockAgentGuidanceRepository,
		mockAgentRepository:         mockAgentRepository,
	}
}

func TestWhenGettingAgentGuidance(t *testing.T) {
	t.Parallel()

	t.Run("Given a guidance row exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentGuidanceRepository.EXPECT().
				GetByAgentId(gomock.Any(), agent.Id).
				Return(&models.AgentGuidance{AgentId: agent.Id, Content: "answer concisely"}, nil)
		}

		t.Run("Should return the content", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "answer concisely", resp.Content)
		})
	})

	t.Run("Given no guidance row exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentGuidanceRepository.EXPECT().GetByAgentId(gomock.Any(), agent.Id).Return(nil, nil)
		}

		t.Run("Should return empty content", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Empty(t, resp.Content)
		})
	})

	t.Run("Given the guidance read fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentGuidanceTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentGuidanceRepository.EXPECT().
				GetByAgentId(gomock.Any(), agent.Id).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return InternalError with guidance_lookup_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "guidance_lookup_failed", internalErr.Code)
		})
	})
}
