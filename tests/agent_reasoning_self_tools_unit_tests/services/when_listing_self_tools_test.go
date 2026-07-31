package services_test

import (
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

type WhenListingSelfToolsTestingSuite struct {
	sut *services.SelfToolRegistry
}

func WhenListingSelfToolsBeforeEach(t *testing.T, toolNames ...string) *WhenListingSelfToolsTestingSuite {
	mockController := gomock.NewController(t)

	selfTools := make([]interfaces.SelfToolInterface, 0, len(toolNames))
	for _, name := range toolNames {
		mockSelfTool := interfaces_mocks.NewMockSelfToolInterface(mockController)
		mockSelfTool.EXPECT().Definition().Return(&models.Tool{Name: name}).AnyTimes()
		selfTools = append(selfTools, mockSelfTool)
	}

	mockSelfToolMetrics := metrics_mocks.NewMockSelfToolMetricsInterface(mockController)
	mockSelfToolMetrics.EXPECT().RecordToolCall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	sut := services.NewSelfToolRegistry(services.NewSelfToolRegistryParams{
		Logger:          zap.NewNop(),
		SelfToolMetrics: mockSelfToolMetrics,
		SelfTools:       selfTools,
	})

	return &WhenListingSelfToolsTestingSuite{sut: sut}
}

func TestWhenListingSelfTools(t *testing.T) {
	t.Parallel()

	t.Run("Given tools registered out of order", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list their definitions sorted by name", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t, "set_guidance", "check_mint", "get_registration")

			tools := suite.sut.ListTools()

			if assert.Len(t, tools, 3) {
				assert.Equal(t, "check_mint", tools[0].Name)
				assert.Equal(t, "get_registration", tools[1].Name)
				assert.Equal(t, "set_guidance", tools[2].Name)
			}
		})
	})

	t.Run("Given two tools register the same name", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the first and list it once", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t, "get_registration", "get_registration")

			tools := suite.sut.ListTools()

			assert.Len(t, tools, 1)
		})
	})

	t.Run("Given no tools are registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list nothing", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolsBeforeEach(t)

			tools := suite.sut.ListTools()

			assert.Empty(t, tools)
		})
	})
}
