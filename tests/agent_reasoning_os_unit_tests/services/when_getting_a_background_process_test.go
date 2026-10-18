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

type WhenGettingABackgroundProcessTestingSuite struct {
	sut *services.ProcessRegistry

	mockController  *gomock.Controller
	osSettings      *os_settings.OsSettings
	commandSettings *os_settings.CommandSettings
}

func WhenGettingABackgroundProcessBeforeEach(t *testing.T) *WhenGettingABackgroundProcessTestingSuite {
	mockController := gomock.NewController(t)

	osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
	commandSettings := os_settings.NewCommandSettings(
		zap.NewNop(),
		koanf.New("."),
		reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")),
	)

	return &WhenGettingABackgroundProcessTestingSuite{
		sut:             services.NewProcessRegistry(commandSettings),
		mockController:  mockController,
		osSettings:      osSettings,
		commandSettings: commandSettings,
	}
}

// handle builds a mock whose Snapshot always reports the given state.
func (s *WhenGettingABackgroundProcessTestingSuite) handle(
	result *os_interfaces.CommandResult,
) *services_mocks.MockCommandHandleInterface {
	mockCommandHandle := services_mocks.NewMockCommandHandleInterface(s.mockController)
	mockCommandHandle.EXPECT().Snapshot().Return(result).AnyTimes()
	return mockCommandHandle
}

func TestWhenGettingABackgroundProcess(t *testing.T) {
	t.Parallel()

	t.Run("Given the session does not exist", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingABackgroundProcessBeforeEach(t)

			_, _, err := suite.sut.Get(uuid.New())

			var notFoundError *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundError)
		})
	})

	t.Run("Given a command that has exited", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fold its exit code into the tracked process", func(t *testing.T) {
			t.Parallel()

			suite := WhenGettingABackgroundProcessBeforeEach(t)
			registered, err := suite.sut.Register(
				"false",
				"/workspace",
				suite.handle(&os_interfaces.CommandResult{ExitCode: lo.ToPtr(1)}),
			)
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
}
