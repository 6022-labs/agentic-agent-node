package services_test

import (
	"testing"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	os_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	services_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/services_mocks/interfaces_mocks"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRegisteringABackgroundProcessTestingSuite struct {
	sut *services.ProcessRegistry

	mockController  *gomock.Controller
	osSettings      *os_settings.OsSettings
	commandSettings *os_settings.CommandSettings
}

func WhenRegisteringABackgroundProcessBeforeEach(t *testing.T) *WhenRegisteringABackgroundProcessTestingSuite {
	mockController := gomock.NewController(t)

	osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
	commandSettings := os_settings.NewCommandSettings(
		zap.NewNop(),
		koanf.New("."),
		reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")),
	)

	return &WhenRegisteringABackgroundProcessTestingSuite{
		sut:             services.NewProcessRegistry(commandSettings),
		mockController:  mockController,
		osSettings:      osSettings,
		commandSettings: commandSettings,
	}
}

// handle builds a mock whose Snapshot always reports the given state.
func (s *WhenRegisteringABackgroundProcessTestingSuite) handle(
	result *os_interfaces.CommandResult,
) *services_mocks.MockCommandHandleInterface {
	mockCommandHandle := services_mocks.NewMockCommandHandleInterface(s.mockController)
	mockCommandHandle.EXPECT().Snapshot().Return(result).AnyTimes()
	return mockCommandHandle
}

func TestWhenRegisteringABackgroundProcess(t *testing.T) {
	t.Parallel()

	t.Run("Given the concurrency ceiling is reached", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the next command", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringABackgroundProcessBeforeEach(t)
			for range suite.commandSettings.MaxBackgroundProcesses {
				_, err := suite.sut.Register(
					"sleep 60",
					"/workspace",
					suite.handle(&os_interfaces.CommandResult{Running: true}),
				)
				assert.NoError(t, err)
			}

			process, err := suite.sut.Register(
				"sleep 60",
				"/workspace",
				suite.handle(&os_interfaces.CommandResult{Running: true}),
			)

			assert.Nil(t, process)
			var conflictError *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictError)
		})

		t.Run("Should free the slot of an exited command", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringABackgroundProcessBeforeEach(t)
			for i := range suite.commandSettings.MaxBackgroundProcesses {
				running := i > 0
				_, err := suite.sut.Register(
					"sleep 60",
					"/workspace",
					suite.handle(&os_interfaces.CommandResult{Running: running}),
				)
				assert.NoError(t, err)
			}

			process, err := suite.sut.Register(
				"echo hi",
				"/workspace",
				suite.handle(&os_interfaces.CommandResult{Running: true}),
			)

			assert.NoError(t, err)
			assert.NotNil(t, process)
		})
	})

	t.Run("Given a started command", func(t *testing.T) {
		t.Parallel()

		t.Run("Should track it as running", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringABackgroundProcessBeforeEach(t)

			process, err := suite.sut.Register(
				"tail -f log",
				"/workspace",
				suite.handle(&os_interfaces.CommandResult{Running: true}),
			)

			assert.NoError(t, err)
			if assert.NotNil(t, process) {
				assert.True(t, process.Running)
				assert.Equal(t, "tail -f log", process.Command)
				assert.Nil(t, process.EndedAt)
			}
			assert.Len(t, suite.sut.List(), 1)
		})
	})
}
