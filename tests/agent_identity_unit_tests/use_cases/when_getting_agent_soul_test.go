package use_cases_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/use_cases"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenGettingAgentSoulTestingSuite struct {
	sut *use_cases.GetAgentSoul

	mockAgentRepository *repository_mocks.MockAgentRepositoryInterface
}

func WhenGettingAgentSoulBeforeEach(t *testing.T) *WhenGettingAgentSoulTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenGettingAgentSoulTestingSuite{
		sut:                 use_cases.NewGetAgentSoul(mockAgentRepository),
		mockAgentRepository: mockAgentRepository,
	}
}

func TestWhenGettingAgentSoul(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should return an InternalError carrying the cause", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var internalErr *apperrors.InternalError
			assert.ErrorAs(t, err, &internalErr)
			assert.Contains(t, internalErr.Message, "db down")
		})
	})

	t.Run("Given no agent row exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should return NotFound with identity_not_found", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.Execute(context.Background())

			var notFoundErr *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundErr)
			assert.Equal(t, "identity_not_found", notFoundErr.Code)
		})
	})

	t.Run("Given a soul exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenGettingAgentSoulTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithSoul("persona text").Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
		}

		t.Run("Should return the soul", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingAgentSoulBeforeEach(t)
			initSuite(suite)

			resp, err := suite.sut.Execute(context.Background())

			assert.NoError(t, err)
			assert.Equal(t, "persona text", resp.Soul)
		})
	})
}
