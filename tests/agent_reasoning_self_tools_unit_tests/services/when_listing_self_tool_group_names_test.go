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

type WhenListingSelfToolGroupNamesTestingSuite struct {
	sut *services.SelfToolRegistry
}

func WhenListingSelfToolGroupNamesBeforeEach(
	t *testing.T,
	groupByName map[string]string,
) *WhenListingSelfToolGroupNamesTestingSuite {
	mockController := gomock.NewController(t)

	selfTools := make([]interfaces.SelfToolInterface, 0, len(groupByName))
	for name, group := range groupByName {
		mockSelfTool := interfaces_mocks.NewMockSelfToolInterface(mockController)
		mockSelfTool.EXPECT().Definition().Return(&models.Tool{Name: name}).AnyTimes()
		mockSelfTool.EXPECT().IsEnabledByDefault().Return(true).AnyTimes()
		mockSelfTool.EXPECT().GroupName().Return(group).AnyTimes()
		selfTools = append(selfTools, mockSelfTool)
	}

	mockSelfToolMetrics := metrics_mocks.NewMockSelfToolMetricsInterface(mockController)
	mockSelfToolMetrics.EXPECT().RecordToolCall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	sut := services.NewSelfToolRegistry(services.NewSelfToolRegistryParams{
		Logger:          zap.NewNop(),
		SelfToolMetrics: mockSelfToolMetrics,
		SelfTools:       selfTools,
	})

	return &WhenListingSelfToolGroupNamesTestingSuite{sut: sut}
}

func TestWhenListingSelfToolGroupNames(t *testing.T) {
	t.Parallel()

	t.Run("Given tools declare their groups", func(t *testing.T) {
		t.Parallel()

		t.Run("Should expose each tool's declared group", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolGroupNamesBeforeEach(t, map[string]string{
				"get_identity": "Identity",
				"run_command":  "Machine",
			})

			assert.Equal(
				t,
				map[string]string{"get_identity": "Identity", "run_command": "Machine"},
				suite.sut.ListGroupNames(),
			)
		})
	})

	t.Run("Given no tools are registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should expose an empty map", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingSelfToolGroupNamesBeforeEach(t, nil)

			assert.Empty(t, suite.sut.ListGroupNames())
		})
	})
}
