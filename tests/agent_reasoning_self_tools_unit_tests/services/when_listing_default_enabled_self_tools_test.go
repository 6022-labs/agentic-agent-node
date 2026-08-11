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

type WhenListingDefaultEnabledSelfToolsTestingSuite struct {
	sut *services.SelfToolRegistry
}

func WhenListingDefaultEnabledSelfToolsBeforeEach(
	t *testing.T,
	defaultEnabledByName map[string]bool,
) *WhenListingDefaultEnabledSelfToolsTestingSuite {
	mockController := gomock.NewController(t)

	selfTools := make([]interfaces.SelfToolInterface, 0, len(defaultEnabledByName))
	for name, enabled := range defaultEnabledByName {
		mockSelfTool := interfaces_mocks.NewMockSelfToolInterface(mockController)
		mockSelfTool.EXPECT().Definition().Return(&models.Tool{Name: name}).AnyTimes()
		mockSelfTool.EXPECT().IsEnabledByDefault().Return(enabled).AnyTimes()
		mockSelfTool.EXPECT().GroupName().Return("General").AnyTimes()
		selfTools = append(selfTools, mockSelfTool)
	}

	mockSelfToolMetrics := metrics_mocks.NewMockSelfToolMetricsInterface(mockController)
	mockSelfToolMetrics.EXPECT().RecordToolCall(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	sut := services.NewSelfToolRegistry(services.NewSelfToolRegistryParams{
		Logger:          zap.NewNop(),
		SelfToolMetrics: mockSelfToolMetrics,
		SelfTools:       selfTools,
	})

	return &WhenListingDefaultEnabledSelfToolsTestingSuite{sut: sut}
}

func TestWhenListingDefaultEnabledSelfTools(t *testing.T) {
	t.Parallel()

	t.Run("Given tools declare different default enablement", func(t *testing.T) {
		t.Parallel()

		t.Run("Should expose each tool's declared default", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingDefaultEnabledSelfToolsBeforeEach(t, map[string]bool{
				"get_identity": true,
				"run_command":  false,
			})

			defaults := suite.sut.ListDefaultEnabled()

			assert.Equal(t, map[string]bool{"get_identity": true, "run_command": false}, defaults)
		})
	})

	t.Run("Given no tools are registered", func(t *testing.T) {
		t.Parallel()

		t.Run("Should expose an empty map", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingDefaultEnabledSelfToolsBeforeEach(t, nil)

			assert.Empty(t, suite.sut.ListDefaultEnabled())
		})
	})
}
