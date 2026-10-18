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
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenKillingABackgroundProcessTestingSuite struct {
	sut *services.ProcessRegistry

	mockController  *gomock.Controller
	osSettings      *os_settings.OsSettings
	commandSettings *os_settings.CommandSettings
}

func WhenKillingABackgroundProcessBeforeEach(t *testing.T) *WhenKillingABackgroundProcessTestingSuite {
	mockController := gomock.NewController(t)

	osSettings := os_settings.NewOsSettings(zap.NewNop(), koanf.New("."))
	commandSettings := os_settings.NewCommandSettings(
		zap.NewNop(),
		koanf.New("."),
		reasoning_settings.NewInferenceLoopSettings(zap.NewNop(), koanf.New(".")),
	)

	return &WhenKillingABackgroundProcessTestingSuite{
		sut:             services.NewProcessRegistry(commandSettings),
		mockController:  mockController,
		osSettings:      osSettings,
		commandSettings: commandSettings,
	}
}

// handle builds a mock whose Snapshot always reports the given state.
func (s *WhenKillingABackgroundProcessTestingSuite) handle(
	result *os_interfaces.CommandResult,
) *services_mocks.MockCommandHandleInterface {
	mockCommandHandle := services_mocks.NewMockCommandHandleInterface(s.mockController)
	mockCommandHandle.EXPECT().Snapshot().Return(result).AnyTimes()
	return mockCommandHandle
}

func TestWhenKillingABackgroundProcess(t *testing.T) {
	t.Parallel()

	t.Run("Given the session does not exist", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a NotFoundError", func(t *testing.T) {
			t.Parallel()

			suite := WhenKillingABackgroundProcessBeforeEach(t)

			err := suite.sut.Kill(uuid.New())

			var notFoundError *apperrors.NotFoundError
			assert.ErrorAs(t, err, &notFoundError)
		})
	})

	t.Run("Given a tracked process", func(t *testing.T) {
		t.Parallel()

		t.Run("Should kill its process group", func(t *testing.T) {
			t.Parallel()

			suite := WhenKillingABackgroundProcessBeforeEach(t)
			mockCommandHandle := suite.handle(&os_interfaces.CommandResult{Running: true})
			mockCommandHandle.EXPECT().Kill().Return(nil)
			registered, err := suite.sut.Register("sleep 60", "/workspace", mockCommandHandle)
			assert.NoError(t, err)

			assert.NoError(t, suite.sut.Kill(registered.Id))
		})
	})
}
