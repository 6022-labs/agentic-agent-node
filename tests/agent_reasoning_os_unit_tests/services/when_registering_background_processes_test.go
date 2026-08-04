package services_test

import (
	"testing"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	os_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	services_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenRegisteringBackgroundProcessesTestingSuite struct {
	sut *services.ProcessRegistry

	mockController  *gomock.Controller
	osSettings      *os_settings.OsSettings
	commandSettings *os_settings.CommandSettings
}

func WhenRegisteringBackgroundProcessesBeforeEach(t *testing.T) *WhenRegisteringBackgroundProcessesTestingSuite {
	mockController := gomock.NewController(t)

	osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
	commandSettings := os_settings.NewCommandSettings(zap.NewNop(), koanf.New("."), reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")))

	return &WhenRegisteringBackgroundProcessesTestingSuite{
		sut:             services.NewProcessRegistry(commandSettings),
		mockController:  mockController,
		osSettings:      osSettings,
		commandSettings: commandSettings,
	}
}

// handle builds a mock whose Snapshot always reports the given state.
func (s *WhenRegisteringBackgroundProcessesTestingSuite) handle(result *os_interfaces.CommandResult) *services_mocks.MockCommandHandleInterface {
	mockCommandHandle := services_mocks.NewMockCommandHandleInterface(s.mockController)
	mockCommandHandle.EXPECT().Snapshot().Return(result).AnyTimes()
	return mockCommandHandle
}

func TestWhenRegisteringBackgroundProcesses(t *testing.T) {
	t.Parallel()

	t.Run("Given a started command", func(t *testing.T) {
		t.Parallel()

		t.Run("Should track it as running", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringBackgroundProcessesBeforeEach(t)

			process, err := suite.sut.Register("tail -f log", "/workspace", suite.handle(&os_interfaces.CommandResult{Running: true}))

			assert.NoError(t, err)
			if assert.NotNil(t, process) {
				assert.True(t, process.Running)
				assert.Equal(t, "tail -f log", process.Command)
				assert.Nil(t, process.EndedAt)
			}
			assert.Len(t, suite.sut.List(), 1)
		})
	})

	t.Run("Given the concurrency ceiling is reached", func(t *testing.T) {
		t.Parallel()

		t.Run("Should refuse the next command", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringBackgroundProcessesBeforeEach(t)
			for range suite.commandSettings.MaxBackgroundProcesses {
				_, err := suite.sut.Register("sleep 60", "/workspace", suite.handle(&os_interfaces.CommandResult{Running: true}))
				assert.NoError(t, err)
			}

			process, err := suite.sut.Register("sleep 60", "/workspace", suite.handle(&os_interfaces.CommandResult{Running: true}))

			assert.Nil(t, process)
			var conflictError *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictError)
		})

		t.Run("Should free a slot once a command exits", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringBackgroundProcessesBeforeEach(t)
			for i := range suite.commandSettings.MaxBackgroundProcesses {
				running := i > 0
				_, err := suite.sut.Register("sleep 60", "/workspace", suite.handle(&os_interfaces.CommandResult{Running: running}))
				assert.NoError(t, err)
			}

			process, err := suite.sut.Register("echo hi", "/workspace", suite.handle(&os_interfaces.CommandResult{Running: true}))

			assert.NoError(t, err)
			assert.NotNil(t, process)
		})
	})

	t.Run("Given a command that has exited", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fold its exit code into the tracked process", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringBackgroundProcessesBeforeEach(t)
			registered, err := suite.sut.Register("false", "/workspace", suite.handle(&os_interfaces.CommandResult{ExitCode: lo.ToPtr(1)}))
			assert.NoError(t, err)

			process, result, err := suite.sut.Get(registered.Id)

			assert.NoError(t, err)
			if assert.NotNil(t, process) {
				assert.False(t, process.Running)
				assert.NotNil(t, process.EndedAt)
				if assert.NotNil(t, process.ExitCode) {
					assert.Equal(t, 1, *process.ExitCode)
				}
			}
			assert.NotNil(t, result)
		})
	})

	t.Run("Given an unknown session id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it as not found", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringBackgroundProcessesBeforeEach(t)

			_, _, getError := suite.sut.Get(uuid.New())
			killError := suite.sut.Kill(uuid.New())

			var notFoundError *apperrors.NotFoundError
			assert.ErrorAs(t, getError, &notFoundError)
			assert.ErrorAs(t, killError, &notFoundError)
		})
	})

	t.Run("Given a tracked process", func(t *testing.T) {
		t.Parallel()

		t.Run("Should kill its process group", func(t *testing.T) {
			t.Parallel()

			suite := WhenRegisteringBackgroundProcessesBeforeEach(t)
			mockCommandHandle := suite.handle(&os_interfaces.CommandResult{Running: true})
			mockCommandHandle.EXPECT().Kill().Return(nil)
			registered, err := suite.sut.Register("sleep 60", "/workspace", mockCommandHandle)
			assert.NoError(t, err)

			assert.NoError(t, suite.sut.Kill(registered.Id))
		})
	})
}
