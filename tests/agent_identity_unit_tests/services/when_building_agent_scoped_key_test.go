package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/services"
	identity_interfaces_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type WhenBuildingAgentScopedKeyTestingSuite struct {
	sut *services.AgentScopedKeyBuilder

	mockAgentResolver *identity_interfaces_mocks.MockAgentResolverInterface
}

func WhenBuildingAgentScopedKeyBeforeEach(t *testing.T) *WhenBuildingAgentScopedKeyTestingSuite {
	mockController := gomock.NewController(t)

	mockAgentResolver := identity_interfaces_mocks.NewMockAgentResolverInterface(mockController)

	return &WhenBuildingAgentScopedKeyTestingSuite{
		sut:               services.NewAgentScopedKeyBuilder(mockAgentResolver),
		mockAgentResolver: mockAgentResolver,
	}
}

func TestWhenBuildingAgentScopedKey(t *testing.T) {
	t.Parallel()

	t.Run("Given the agent lookup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingAgentScopedKeyTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(nil, errors.New("db down"))
		}

		t.Run("Should propagate the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAgentScopedKeyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ScopedKey(context.Background(), "wallet")

			assert.ErrorContains(t, err, "db down")
		})
	})

	t.Run("Given no agent row exists", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenBuildingAgentScopedKeyTestingSuite) {
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should error", func(t *testing.T) {
			t.Parallel()

			suite := WhenBuildingAgentScopedKeyBeforeEach(t)
			initSuite(suite)

			_, err := suite.sut.ScopedKey(context.Background(), "wallet")

			assert.ErrorContains(t, err, "no agent record exists yet")
		})
	})

	t.Run("Given an agent row exists", func(t *testing.T) {
		t.Parallel()

		id := uuid.New()

		initSuite := func(suite *WhenBuildingAgentScopedKeyTestingSuite) {
			agent := model_builders.NewAgentBuilder().WithId(id).Build()
			suite.mockAgentResolver.EXPECT().Get(gomock.Any()).Return(agent, nil).Times(1)
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
}
