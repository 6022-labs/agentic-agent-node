package services_test

import (
	"context"
	"testing"

	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/services"
	os_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning_os/settings"
	apperrors "github.com/6022-labs/agentic-agent-node/src/common/errors"
	metrics_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/metrics_mocks/interfaces_mocks"
	services_mocks "github.com/6022-labs/agentic-agent-node/tests/agent_reasoning_os_mocks/services_mocks/interfaces_mocks"
	"github.com/google/uuid"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

type WhenKillingACommandTestingSuite struct {
	sut *services.CommandExecutor

	osSettings          *os_settings.OsSettings
	commandSettings     *os_settings.CommandSettings
	mockCommandRunner   *services_mocks.MockCommandRunnerInterface
	mockCommandHandle   *services_mocks.MockCommandHandleInterface
	mockProcessRegistry *services_mocks.MockProcessRegistryInterface
	mockCommandMetrics  *metrics_mocks.MockCommandMetricsInterface
}

func WhenKillingACommandBeforeEach(t *testing.T) *WhenKillingACommandTestingSuite {
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

	return &WhenKillingACommandTestingSuite{
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

func TestWhenKillingACommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a nil session id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a ValidationError without touching dependencies", func(t *testing.T) {
			t.Parallel()

			suite := WhenKillingACommandBeforeEach(t)

			err := suite.sut.Kill(ctx, uuid.Nil)

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, err, &validationError)
		})
	})

	t.Run("Given a tracked process", func(t *testing.T) {
		t.Parallel()

		sessionId := uuid.New()
		initSuite := func(suite *WhenKillingACommandTestingSuite) {
			suite.mockProcessRegistry.EXPECT().Kill(sessionId).Return(nil)
			suite.mockCommandMetrics.EXPECT().RecordProcessKilled(ctx)
		}

		t.Run("Should kill it through the registry", func(t *testing.T) {
			t.Parallel()

			suite := WhenKillingACommandBeforeEach(t)
			initSuite(suite)

			assert.NoError(t, suite.sut.Kill(ctx, sessionId))
		})

		t.Run("Should record the kill metric", func(t *testing.T) {
			t.Parallel()

			suite := WhenKillingACommandBeforeEach(t)
			initSuite(suite)
			_ = suite.sut.Kill(ctx, sessionId)
		})
	})
}
