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
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenSettingAgentGuidanceTestingSuite struct {
	sut *use_cases.SetAgentGuidance

	mockAgentGuidanceRepository *repositories_mocks.MockAgentGuidanceRepositoryInterface
	mockAgentResolver           *identity_interfaces_mocks.MockAgentResolverInterface
}

func WhenSettingAgentGuidanceBeforeEach(t *testing.T) *WhenSettingAgentGuidanceTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentGuidanceRepository := repositories_mocks.NewMockAgentGuidanceRepositoryInterface(mockController)
	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)

	return &WhenSettingAgentGuidanceTestingSuite{
		sut:                         use_cases.NewSetAgentGuidance(mockAgentGuidanceRepository, mockAgentResolver),
		mockAgentGuidanceRepository: mockAgentGuidanceRepository,
		mockAgentResolver:           mockAgentResolver,
	}
}

func TestWhenSettingAgentGuidance(t *testing.T) {
	t.Parallel()

	t.Run("Given saving the agent guidance fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentGuidanceRepository.EXPECT().
				Upsert(gomock.Any(), gomock.Any()).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should return InternalError with guidance_update_failed", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background(), &requests.SetAgentGuidanceRequest{Content: "x"})

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Equal(t, "guidance_update_failed", internalErr.Code)
		})
	})

	t.Run("Given the agent exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSettingAgentGuidanceTestingSuite) {
			agent := model_builders.NewAgentBuilder().Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentGuidanceRepository.EXPECT().
				Upsert(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, g *models.AgentGuidance) (*models.AgentGuidance, error) {
					assert.Equal(t, agent.Id, g.AgentId)
					assert.Equal(t, "answer concisely", g.Content)
					return g, nil
				})
		}

		t.Run("Should upsert the guidance for the agent", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(
				context.Background(),
				&requests.SetAgentGuidanceRequest{Content: "answer concisely"},
			)

			assert.NoError(t, err)
		})

		t.Run("Should echo the guidance", func(t *testing.T) {
			t.Parallel()

			suite := WhenSettingAgentGuidanceBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(
				context.Background(),
				&requests.SetAgentGuidanceRequest{Content: "answer concisely"},
			)

			assert.NoError(t, err)
			assert.Equal(t, "answer concisely", resp.Content)
		})
	})
}
