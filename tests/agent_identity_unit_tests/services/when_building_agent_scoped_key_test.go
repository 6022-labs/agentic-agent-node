package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	repository_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenBuildingAgentScopedKeyTestingSuite struct {
	sut *services.AgentScopedKeyBuilder

	mockAgentRepository *repository_mocks.MockAgentRepositoryInterface
}

func WhenBuildingAgentScopedKeyBeforeEach(t *testing.T) *WhenBuildingAgentScopedKeyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentRepository := repository_mocks.NewMockAgentRepositoryInterface(mockController)

	return &WhenBuildingAgentScopedKeyTestingSuite{
		sut:                 services.NewAgentScopedKeyBuilder(mockAgentRepository),
		mockAgentRepository: mockAgentRepository,
	}
}

func TestWhenBuildingAgentScopedKey(t *testing.T) {
	t.Parallel()

	t.Run("Given an agent record exists", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenBuildingAgentScopedKeyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithId(id).Build()
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
		}

		t.Run("Should namespace the key under agents/<id>", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAgentScopedKeyBeforeEach(t)
			initSuite(suite)

			scopedKey, err := suite.sut.ScopedKey(context.Background(), "wallet")

			assert.NoError(t, err)
			assert.Equal(t, "agents/"+id.String()+"/wallet", scopedKey)
		})

		t.Run("Should cache the id and query the repository only once", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAgentScopedKeyBeforeEach(t)
			initSuite(suite)

			first, err := suite.sut.ScopedKey(context.Background(), "llm_secret")
			assert.NoError(t, err)

			second, err := suite.sut.ScopedKey(context.Background(), "wallet")
			assert.NoError(t, err)

			assert.Equal(t, "agents/"+id.String()+"/llm_secret", first)
			assert.Equal(t, "agents/"+id.String()+"/wallet", second)
		})
	})

	t.Run("Given no agent record exists yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingAgentScopedKeyTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should error", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAgentScopedKeyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ScopedKey(context.Background(), "wallet")

			assert.Error(t, err)
		})
	})

	t.Run("Given the repository lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingAgentScopedKeyTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAgentScopedKeyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ScopedKey(context.Background(), "wallet")

			assert.ErrorContains(t, err, "db down")
		})
	})
}
