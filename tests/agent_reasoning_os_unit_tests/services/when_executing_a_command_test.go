package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

type WhenExecutingACommandTestingSuite struct {
	sut *services.CommandExecutor

	osSettings          *os_settings.OsSettings
	commandSettings     *os_settings.CommandSettings
	mockCommandRunner   *services_mocks.MockCommandRunnerInterface
	mockCommandHandle   *services_mocks.MockCommandHandleInterface
	mockProcessRegistry *services_mocks.MockProcessRegistryInterface
	mockCommandMetrics  *metrics_mocks.MockCommandMetricsInterface
}

func WhenExecutingACommandBeforeEach(t *testing.T) *WhenExecutingACommandTestingSuite {
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

	return &WhenExecutingACommandTestingSuite{
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

func TestWhenExecutingACommand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("Given a foreground command", func(t *testing.T) {
		t.Parallel()

		t.Run("Should run it with the default workdir and timeout", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			suite.mockCommandRunner.EXPECT().
				Run(ctx, &os_interfaces.CommandSpec{
					Command: "echo hi",
					Workdir: suite.osSettings.Workdir,
					Timeout: time.Duration(suite.commandSettings.DefaultTimeoutSeconds) * time.Second,
				}).
				Return(&os_interfaces.CommandResult{Stdout: "hi\n"}, nil)

			result, err := suite.sut.Run(ctx, "echo hi", "", nil, nil, nil)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, "hi\n", result.Stdout)
				assert.Nil(t, result.SessionId)
			}
		})
	})

	t.Run("Given a timeout above the ceiling", func(t *testing.T) {
		t.Parallel()

		t.Run("Should clamp it instead of failing", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			suite.mockCommandRunner.EXPECT().
				Run(ctx, &os_interfaces.CommandSpec{
					Command: "sleep 9000",
					Workdir: "/tmp",
					Timeout: time.Duration(suite.commandSettings.MaxTimeoutSeconds) * time.Second,
				}).
				Return(&os_interfaces.CommandResult{TimedOut: true, ExitCode: lo.ToPtr(-1)}, nil)

			result, err := suite.sut.Run(ctx, "sleep 9000", "/tmp", lo.ToPtr(10_000_000_000), nil, nil)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.True(t, result.TimedOut)
			}
		})
	})

	t.Run("Given a background command", func(t *testing.T) {
		t.Parallel()

		t.Run("Should register it and return its session id without waiting", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			sessionId := uuid.New()
			suite.mockCommandRunner.EXPECT().Start(gomock.Any()).Return(suite.mockCommandHandle, nil)
			suite.mockProcessRegistry.EXPECT().
				Register("tail -f log", suite.osSettings.Workdir, suite.mockCommandHandle).
				Return(&os_interfaces.BackgroundProcess{Id: sessionId, Running: true}, nil)
			suite.mockCommandMetrics.EXPECT().RecordBackgroundStarted(ctx)
			suite.mockCommandHandle.EXPECT().Snapshot().Return(&os_interfaces.CommandResult{Running: true})

			result, err := suite.sut.Run(ctx, "tail -f log", "", nil, nil, lo.ToPtr(true))

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.True(t, result.Running)
				if assert.NotNil(t, result.SessionId) {
					assert.Equal(t, sessionId, *result.SessionId)
				}
			}
		})

		t.Run("Should report the exit code of a command that already finished", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			sessionId := uuid.New()
			suite.mockCommandRunner.EXPECT().Start(gomock.Any()).Return(suite.mockCommandHandle, nil)
			suite.mockProcessRegistry.EXPECT().
				Register(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(&os_interfaces.BackgroundProcess{Id: sessionId}, nil)
			suite.mockCommandMetrics.EXPECT().RecordBackgroundStarted(ctx)
			suite.mockCommandHandle.EXPECT().
				Snapshot().
				Return(&os_interfaces.CommandResult{ExitCode: lo.ToPtr(127), Stderr: "not found"})

			result, err := suite.sut.Run(ctx, "npm ci", "", nil, nil, lo.ToPtr(true))

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.False(t, result.Running)
				if assert.NotNil(t, result.ExitCode) {
					assert.Equal(t, 127, *result.ExitCode)
				}
				assert.Equal(t, "not found", result.Stderr)
			}
		})

		t.Run(
			"Should let a background command run to the background ceiling, not the foreground default",
			func(t *testing.T) {
				t.Parallel()

				suite := WhenExecutingACommandBeforeEach(t)
				suite.mockCommandRunner.EXPECT().
					Start(&os_interfaces.CommandSpec{
						Command: "tail -f log",
						Workdir: suite.osSettings.Workdir,
						Timeout: time.Duration(suite.commandSettings.MaxBackgroundSeconds) * time.Second,
					}).
					Return(suite.mockCommandHandle, nil)
				suite.mockProcessRegistry.EXPECT().
					Register(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(&os_interfaces.BackgroundProcess{Id: uuid.New(), Running: true}, nil)
				suite.mockCommandMetrics.EXPECT().RecordBackgroundStarted(ctx)
				suite.mockCommandHandle.EXPECT().Snapshot().Return(&os_interfaces.CommandResult{Running: true})

				_, err := suite.sut.Run(ctx, "tail -f log", "", nil, nil, lo.ToPtr(true))

				assert.NoError(t, err)
			},
		)

		t.Run("Should kill the process when the registry refuses it", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			suite.mockCommandRunner.EXPECT().Start(gomock.Any()).Return(suite.mockCommandHandle, nil)
			suite.mockProcessRegistry.EXPECT().
				Register(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewConflictError("too_many_background_processes", "full"))
			suite.mockCommandHandle.EXPECT().Kill().Return(nil)

			result, err := suite.sut.Run(ctx, "tail -f log", "", nil, nil, lo.ToPtr(true))

			assert.Nil(t, result)
			var conflictError *apperrors.ConflictError
			assert.ErrorAs(t, err, &conflictError)
		})
	})

	t.Run("Given invalid arguments", func(t *testing.T) {
		t.Parallel()

		t.Run("Should fail validation without running anything", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)

			_, emptyCommandError := suite.sut.Run(ctx, "  ", "", nil, nil, nil)
			_, zeroTimeoutError := suite.sut.Run(ctx, "echo hi", "", lo.ToPtr(0), nil, nil)
			_, negativeYieldError := suite.sut.Run(ctx, "echo hi", "", nil, lo.ToPtr(-1), nil)
			_, nilOutputError := suite.sut.ReadOutput(uuid.Nil)
			nilKillError := suite.sut.Kill(ctx, uuid.Nil)

			var validationError *apperrors.ValidationError
			assert.ErrorAs(t, emptyCommandError, &validationError)
			assert.ErrorAs(t, zeroTimeoutError, &validationError)
			assert.ErrorAs(t, negativeYieldError, &validationError)
			assert.ErrorAs(t, nilOutputError, &validationError)
			assert.ErrorAs(t, nilKillError, &validationError)
		})
	})

	t.Run("Given the runner fails", func(t *testing.T) {
		t.Parallel()

		t.Run("Should surface an internal error", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			suite.mockCommandRunner.EXPECT().Run(ctx, gomock.Any()).Return(nil, errors.New("fork failed"))

			result, err := suite.sut.Run(ctx, "echo hi", "", nil, nil, nil)

			assert.Nil(t, result)
			var internalError *apperrors.InternalError
			assert.ErrorAs(t, err, &internalError)
		})
	})

	t.Run("Given a tracked process", func(t *testing.T) {
		t.Parallel()

		t.Run("Should stamp the session id onto the output it reads", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			sessionId := uuid.New()
			suite.mockProcessRegistry.EXPECT().
				Get(sessionId).
				Return(&os_interfaces.BackgroundProcess{Id: sessionId}, &os_interfaces.CommandResult{Stdout: "done", ExitCode: lo.ToPtr(3)}, nil)

			result, err := suite.sut.ReadOutput(sessionId)

			assert.NoError(t, err)
			if assert.NotNil(t, result) && assert.NotNil(t, result.SessionId) {
				assert.Equal(t, sessionId, *result.SessionId)
				if assert.NotNil(t, result.ExitCode) {
					assert.Equal(t, 3, *result.ExitCode)
				}
			}
		})

		t.Run("Should kill it and record the metric", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			sessionId := uuid.New()
			suite.mockProcessRegistry.EXPECT().Kill(sessionId).Return(nil)
			suite.mockCommandMetrics.EXPECT().RecordProcessKilled(ctx)

			assert.NoError(t, suite.sut.Kill(ctx, sessionId))
		})
	})

	t.Run("Given tracked processes in any order", func(t *testing.T) {
		t.Parallel()

		t.Run("Should list them oldest first", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
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

	t.Run("Given a foreground command with a yield delay", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its result directly when it finishes before yielding", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)

			done := make(chan struct{})
			close(done)

			suite.mockCommandRunner.EXPECT().Start(gomock.Any()).Return(suite.mockCommandHandle, nil)
			suite.mockCommandHandle.EXPECT().Done().Return(done)
			suite.mockCommandHandle.EXPECT().
				Snapshot().
				Return(&os_interfaces.CommandResult{Stdout: "done\n", ExitCode: lo.ToPtr(0)})

			result, err := suite.sut.Run(ctx, "echo done", "", lo.ToPtr(60), lo.ToPtr(30), nil)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Nil(t, result.SessionId)
				assert.Equal(t, "done\n", result.Stdout)
			}
		})

		t.Run("Should detach it with its output so far once the delay passes", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)
			sessionId := uuid.New()

			suite.mockCommandRunner.EXPECT().Start(gomock.Any()).Return(suite.mockCommandHandle, nil)
			suite.mockCommandHandle.EXPECT().Done().Return(make(chan struct{}))
			suite.mockProcessRegistry.EXPECT().
				Register("sleep 60", suite.osSettings.Workdir, suite.mockCommandHandle).
				Return(&os_interfaces.BackgroundProcess{Id: sessionId, Running: true}, nil)
			suite.mockCommandMetrics.EXPECT().RecordBackgroundStarted(ctx)
			suite.mockCommandHandle.EXPECT().
				Snapshot().
				Return(&os_interfaces.CommandResult{Stdout: "partial", Running: true})

			result, err := suite.sut.Run(ctx, "sleep 60", "", lo.ToPtr(60), lo.ToPtr(1), nil)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Equal(t, &sessionId, result.SessionId)
				assert.Equal(t, "partial", result.Stdout)
				assert.True(t, result.Running)
			}
		})

		t.Run("Should wait it out when the registry is full, so it is never left untracked", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)

			exited := make(chan struct{})
			close(exited)

			suite.mockCommandRunner.EXPECT().Start(gomock.Any()).Return(suite.mockCommandHandle, nil)
			gomock.InOrder(
				suite.mockCommandHandle.EXPECT().Done().Return(make(chan struct{})),
				suite.mockCommandHandle.EXPECT().Done().Return(exited),
			)
			suite.mockProcessRegistry.EXPECT().
				Register(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, apperrors.NewConflictError("too_many_background_processes", "full"))
			suite.mockCommandHandle.EXPECT().
				Snapshot().
				Return(&os_interfaces.CommandResult{ExitCode: lo.ToPtr(0)})

			result, err := suite.sut.Run(ctx, "sleep 60", "", lo.ToPtr(60), lo.ToPtr(1), nil)

			assert.NoError(t, err)
			if assert.NotNil(t, result) {
				assert.Nil(t, result.SessionId)
				assert.Equal(t, lo.ToPtr(0), result.ExitCode)
			}
		})

		t.Run("Should wait the command out when the delay is explicitly zero", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)

			suite.mockCommandRunner.EXPECT().
				Run(ctx, gomock.Any()).
				Return(&os_interfaces.CommandResult{ExitCode: lo.ToPtr(0)}, nil)

			result, err := suite.sut.Run(ctx, "echo hi", "", nil, lo.ToPtr(0), nil)

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})

		t.Run("Should skip yielding when the delay reaches the timeout", func(t *testing.T) {
			t.Parallel()

			suite := WhenExecutingACommandBeforeEach(t)

			suite.mockCommandRunner.EXPECT().
				Run(ctx, gomock.Any()).
				Return(&os_interfaces.CommandResult{ExitCode: lo.ToPtr(0)}, nil)

			result, err := suite.sut.Run(ctx, "echo hi", "", lo.ToPtr(30), lo.ToPtr(30), nil)

			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	})
}
