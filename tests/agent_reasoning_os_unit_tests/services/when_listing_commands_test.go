package services_test

import (
	"testing"
	"time"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	os_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/metrics_mocks/interfaces_mocks"
	services_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenListingCommandsTestingSuite struct {
	sut *services.CommandExecutor

	osSettings          *os_settings.OsSettings
	commandSettings     *os_settings.CommandSettings
	mockCommandRunner   *services_mocks.MockCommandRunnerInterface
	mockCommandHandle   *services_mocks.MockCommandHandleInterface
	mockProcessRegistry *services_mocks.MockProcessRegistryInterface
	mockCommandMetrics  *metrics_mocks.MockCommandMetricsInterface
}

func WhenListingCommandsBeforeEach(t *testing.T) *WhenListingCommandsTestingSuite {
	mockController := gomock.NewController(t)

	k := koanf.New(".")
	_ = k.Load(confmap.Provider(map[string]any{"command.default_yield_seconds": 0}, "."), nil)

	osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
	commandSettings := os_settings.NewCommandSettings(
		zap.NewNop(),
		k,
		reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")),
	)
	mockCommandRunner := services_mocks.NewMockCommandRunnerInterface(mockController)
	mockCommandHandle := services_mocks.NewMockCommandHandleInterface(mockController)
	mockProcessRegistry := services_mocks.NewMockProcessRegistryInterface(mockController)
	mockCommandMetrics := metrics_mocks.NewMockCommandMetricsInterface(mockController)

	return &WhenListingCommandsTestingSuite{
		sut: services.NewCommandExecutor(
			mockCommandRunner,
			mockProcessRegistry,
			osSettings,
			commandSettings,
			mockCommandMetrics,
		),
		osSettings:          osSettings,
		commandSettings:     commandSettings,
		mockCommandRunner:   mockCommandRunner,
		mockCommandHandle:   mockCommandHandle,
		mockProcessRegistry: mockProcessRegistry,
		mockCommandMetrics:  mockCommandMetrics,
	}
}

func TestWhenListingCommands(t *testing.T) {
	t.Parallel()

	t.Run("Given tracked processes in any order", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list them oldest first", func(t *testing.T) {
			t.Parallel()

			suite := WhenListingCommandsBeforeEach(t)
			startedAt := time.Now().UTC()
			suite.mockProcessRegistry.EXPECT().List().Return([]*os_interfaces.BackgroundProcess{
				{Id: uuid.New(), Command: "second", StartedAt: startedAt.Add(time.Minute)},
				{Id: uuid.New(), Command: "first", StartedAt: startedAt},
			})

			processes := suite.sut.List()

			if assert.Len(t, processes, 2) {
				assert.Equal(t, "first", processes[0].Command)
				assert.Equal(t, "second", processes[1].Command)
			}
		})
	})
}
