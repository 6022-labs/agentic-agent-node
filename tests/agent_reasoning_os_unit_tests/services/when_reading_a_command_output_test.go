package services_test

import (
	"testing"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	os_interfaces "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services/interfaces"
	os_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/metrics_mocks/interfaces_mocks"
	services_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenReadingACommandOutputTestingSuite struct {
	sut *services.CommandExecutor

	osSettings          *os_settings.OsSettings
	commandSettings     *os_settings.CommandSettings
	mockCommandRunner   *services_mocks.MockCommandRunnerInterface
	mockCommandHandle   *services_mocks.MockCommandHandleInterface
	mockProcessRegistry *services_mocks.MockProcessRegistryInterface
	mockCommandMetrics  *metrics_mocks.MockCommandMetricsInterface
}

func WhenReadingACommandOutputBeforeEach(t *testing.T) *WhenReadingACommandOutputTestingSuite {
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

	return &WhenReadingACommandOutputTestingSuite{
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

func TestWhenReadingACommandOutput(t *testing.T) {
	t.Parallel()

	t.Run("Given a nil session id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingACommandOutputBeforeEach(t)

			_, err := suite.sut.ReadOutput(uuid.Nil)

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
		})
	})

	t.Run("Given a tracked process", func(t *testing.T) {
		t.Parallel()

		t.Run("Should stamp the session id onto the output it reads", func(t *testing.T) {
			t.Parallel()

			suite := WhenReadingACommandOutputBeforeEach(t)
			sessionId := uuid.New()
			suite.mockProcessRegistry.EXPECT().
				Get(sessionId).
				Return(
					&os_interfaces.BackgroundProcess{Id: sessionId},
					&os_interfaces.CommandResult{Stdout: "done", ExitCode: lo.ToPtr(3)},
					nil,
				)

			result, err := suite.sut.ReadOutput(sessionId)

			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.NotNil(t, result.SessionId) {
				assert.Equal(t, sessionId, *result.SessionId)
				if assert.NotNil(t, result.ExitCode) {
					assert.Equal(t, 3, *result.ExitCode)
				}
			}
		})
	})
}
