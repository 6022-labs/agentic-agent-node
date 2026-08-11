package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services"
	identity_repositories_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_identity_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/repositories_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/model_builders"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReconcilingSelfToolsTestingSuite struct {
	sut *services.SelfToolReconciler

	mockSelfToolRegistry   *interfaces_mocks.MockSelfToolRegistryInterface
	mockSelfToolRepository *repositories_mocks.MockSelfToolRepositoryInterface
	mockAgentRepository    *identity_repositories_mocks.MockAgentRepositoryInterface
}

func WhenReconcilingSelfToolsBeforeEach(t *testing.T) *WhenReconcilingSelfToolsTestingSuite {
	mockController := gomock.NewController(t)

	mockSelfToolRegistry := interfaces_mocks.NewMockSelfToolRegistryInterface(mockController)
	mockSelfToolRepository := repositories_mocks.NewMockSelfToolRepositoryInterface(mockController)
	mockAgentRepository := identity_repositories_mocks.NewMockAgentRepositoryInterface(mockController)

	sut := services.NewSelfToolReconciler(
		zap.NewNop(),
		mockSelfToolRegistry,
		mockSelfToolRepository,
		mockAgentRepository,
	)

	return &WhenReconcilingSelfToolsTestingSuite{
		sut: sut,

		mockSelfToolRegistry:   mockSelfToolRegistry,
		mockSelfToolRepository: mockSelfToolRepository,
		mockAgentRepository:    mockAgentRepository,
	}
}

func TestWhenReconcilingSelfTools(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	agent := model_builders.NewAgentBuilder().Build()

	t.Run("Given declared self tools have no rows yet", func(t *testing.T) {
		t.Parallel()

		defaults := map[string]bool{"get_identity": true, "run_command": false}
		created := map[string]bool{}

		initSuite := func(suite *WhenReconcilingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockSelfToolRegistry.EXPECT().ListDefaultEnabled().Return(defaults)
			suite.mockSelfToolRepository.EXPECT().
				ListExistingNames(gomock.Any(), agent.Id, gomock.InAnyOrder([]string{"get_identity", "run_command"})).
				Return(nil, nil)
			suite.mockSelfToolRepository.EXPECT().Create(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, tool *models.SelfTool) (*models.SelfTool, error) {
					assert.Equal(t, agent.Id, tool.AgentId)
					created[tool.Name] = tool.Enabled
					return tool, nil
				}).Times(2)
			suite.mockSelfToolRepository.EXPECT().
				DeleteAllExceptNames(gomock.Any(), agent.Id, gomock.InAnyOrder([]string{"get_identity", "run_command"})).
				Return(int64(0), nil)
		}

		t.Run("Should insert every missing tool with its default enabled value", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingSelfToolsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reconcile(ctx)

			assert.NoError(t, err)
			assert.Equal(t, defaults, created)
		})
	})

	t.Run("Given stale rows exist in the database", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockSelfToolRegistry.EXPECT().ListDefaultEnabled().Return(map[string]bool{"get_identity": true})
			suite.mockSelfToolRepository.EXPECT().
				ListExistingNames(gomock.Any(), agent.Id, []string{"get_identity"}).
				Return([]string{"get_identity"}, nil)
			// No Create expectation — the stale cleanup happens set-based, without fetching rows.
			suite.mockSelfToolRepository.EXPECT().
				DeleteAllExceptNames(gomock.Any(), agent.Id, []string{"get_identity"}).
				Return(int64(1), nil)
		}

		t.Run("Should delete them with a single set-based call", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingSelfToolsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reconcile(ctx)

			assert.NoError(t, err)
		})
	})

	t.Run("Given every row matches a declared self tool", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockSelfToolRegistry.EXPECT().ListDefaultEnabled().Return(map[string]bool{"get_identity": true})
			suite.mockSelfToolRepository.EXPECT().
				ListExistingNames(gomock.Any(), agent.Id, []string{"get_identity"}).
				Return([]string{"get_identity"}, nil)
			// No Create expectation — inserting a matching row would fail the test.
			suite.mockSelfToolRepository.EXPECT().
				DeleteAllExceptNames(gomock.Any(), agent.Id, []string{"get_identity"}).
				Return(int64(0), nil)
		}

		t.Run("Should not insert anything", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingSelfToolsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reconcile(ctx)

			assert.NoError(t, err)
		})
	})

	t.Run("Given the agent is not initialized yet", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(nil, nil)
		}

		t.Run("Should fail without touching the repository", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingSelfToolsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reconcile(ctx)

			assert.Error(t, err)
			assert.ErrorContains(t, err, "not initialized")
		})
	})

	t.Run("Given listing the existing names fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockSelfToolRegistry.EXPECT().ListDefaultEnabled().Return(map[string]bool{"get_identity": true})
			suite.mockSelfToolRepository.EXPECT().
				ListExistingNames(gomock.Any(), agent.Id, []string{"get_identity"}).
				Return(nil, errors.New("db down"))
		}

		t.Run("Should wrap the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingSelfToolsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reconcile(ctx)

			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to list existing self tool names")
		})
	})

	t.Run("Given the stale cleanup fails", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenReconcilingSelfToolsTestingSuite) {
			suite.mockAgentRepository.EXPECT().Get(gomock.Any()).Return(agent, nil)
			suite.mockSelfToolRegistry.EXPECT().ListDefaultEnabled().Return(map[string]bool{"get_identity": true})
			suite.mockSelfToolRepository.EXPECT().
				ListExistingNames(gomock.Any(), agent.Id, []string{"get_identity"}).
				Return([]string{"get_identity"}, nil)
			suite.mockSelfToolRepository.EXPECT().
				DeleteAllExceptNames(gomock.Any(), agent.Id, []string{"get_identity"}).
				Return(int64(0), errors.New("db down"))
		}

		t.Run("Should wrap the error", func(t *testing.T) {
			t.Parallel()

			suite := WhenReconcilingSelfToolsBeforeEach(t)
			initSuite(suite)

			err := suite.sut.Reconcile(ctx)

			assert.Error(t, err)
			assert.ErrorContains(t, err, "failed to delete stale self tool rows")
		})
	})
}
