package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/models"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_self_tools/services"
	"github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_mocks/services_mocks/interfaces_mocks"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_self_tools_mocks/metrics_mocks/interfaces_mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenCallingSelfToolTestingSuite struct {
	sut *services.SelfToolRegistry

	mockSelfTool *interfaces_mocks.MockSelfToolInterface
}

func WhenCallingSelfToolBeforeEach(t *testing.T) *WhenCallingSelfToolTestingSuite {
	mockController := gomock.NewController(t)

	mockSelfTool := interfaces_mocks.NewMockSelfToolInterface(mockController)
	mockSelfTool.EXPECT().
		Definition().
		Return(&models.Tool{Name: "get_registration", Description: "read own registration"}).
		AnyTimes()
	mockSelfTool.EXPECT().IsEnabledByDefault().Return(true).AnyTimes()

	mockSelfToolMetrics := metrics_mocks.NewMockSelfToolMetricsInterface(mockController)
	mockSelfToolMetrics.EXPECT().RecordToolCall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	sut := services.NewSelfToolRegistry(services.NewSelfToolRegistryParams{
		Logger:          zap.NewNop(),
		SelfToolMetrics: mockSelfToolMetrics,
		SelfTools:       []interfaces.SelfToolInterface{mockSelfTool},
	})

	return &WhenCallingSelfToolTestingSuite{
		sut:          sut,
		mockSelfTool: mockSelfTool,
	}
}

func TestWhenCallingSelfTool(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a registered tool succeeds", func(t *testing.T) {
		t.Parallel()

		args := map[string]any{"key": "value"}
		result := map[string]any{"status": "minted"}

		initSuite := func(suite *WhenCallingSelfToolTestingSuite) {
			suite.mockSelfTool.EXPECT().Execute(ctx, args).Return(result, nil)
		}

		t.Run("Should return the tool result", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingSelfToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.CallTool(ctx, "get_registration", args)

			assert.NoError(t, err)
			assert.Equal(t, result, out)
		})
	})

	t.Run("Given a registered tool fails", func(t *testing.T) {
		t.Parallel()

		toolErr := errors.New("registration not found")
		initSuite := func(suite *WhenCallingSelfToolTestingSuite) {
			suite.mockSelfTool.EXPECT().Execute(ctx, gomock.Nil()).Return(nil, toolErr)
		}

		t.Run("Should propagate the error for the caller to surface", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingSelfToolBeforeEach(t)
			initSuite(suite)

			out, err := suite.sut.CallTool(ctx, "get_registration", nil)

			assert.Nil(t, out)
			assert.ErrorIs(t, err, toolErr)
		})
	})

	t.Run("Given the tool name is unknown", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail", func(t *testing.T) {
			t.Parallel()

			suite := WhenCallingSelfToolBeforeEach(t)

			out, err := suite.sut.CallTool(ctx, "ghost", nil)

			assert.Nil(t, out)
			assert.ErrorContains(t, err, "unknown self tool")
		})
	})
}
