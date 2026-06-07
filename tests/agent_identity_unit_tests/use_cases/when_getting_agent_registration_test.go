package use_cases_test

import (
	"context"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentRegistrationTestingSuite struct {
	sut *use_cases.GetAgentRegistration

	mockAgentRepository                  *repository_mocks.MockAgentRepositoryInterface
	mockAgentRegistrationStateRepository *repository_mocks.MockAgentRegistrationStateRepositoryInterface
}

func WhenGettingAgentRegistrationBeforeEach(t *testing.T) *WhenGettingAgentRegistrationTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)
	mockAgentRegistrationStateRepository := repository_mocks.NewMockAgentRegistrationStateRepositoryInterface(mockController)

	sut := use_cases.NewGetAgentRegistration(mockAgentRepository, mockAgentRegistrationStateRepository)

	return &WhenGettingAgentRegistrationTestingSuite{
		sut:                                  sut,
		mockAgentRepository:                  mockAgentRepository,
		mockAgentRegistrationStateRepository: mockAgentRegistrationStateRepository,
	}
}

func TestWhenGettingAgentRegistration(t *testing.T) {
	t.Parallel()

	t.Run("Given a saved draft", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentRegistrationTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithName("alice").Build()
			state := model_builders.NewAgentRegistrationStateBuilder().Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockAgentRegistrationStateRepository.EXPECT().Get(gomock.Any()).Return(state, nil)
		}

		t.Run("Should return the stored draft and status", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.NotNil(t, resp)
		})
	})

	t.Run("Given no draft has been saved", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentRegistrationTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return a NotFound", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentRegistrationBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
		})
	})
}
