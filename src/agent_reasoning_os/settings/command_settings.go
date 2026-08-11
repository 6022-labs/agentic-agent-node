package settings

import (
	reasoning_settings "github.com/6022-labs/agentic-agent-node/src/agent_reasoning/settings"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const CommandSettingsKey = "command"

// CommandSettings bounds the commands the agent runs. MaxTimeoutSeconds defaults below the inference loop's own
// ceiling, since a command that outlives the loop takes the whole turn down with it.
type CommandSettings struct {
	DefaultTimeoutSeconds    int `koanf:"default_timeout_seconds"`
	MaxTimeoutSeconds        int `koanf:"max_timeout_seconds"`
	DefaultYieldSeconds      int `koanf:"default_yield_seconds"`
	MaxOutputChars           int `koanf:"max_output_chars"`
	MaxBackgroundProcesses   int `koanf:"max_background_processes"`
	MaxBackgroundSeconds     int `koanf:"max_background_seconds"`
	FinishedRetentionSeconds int `koanf:"finished_retention_seconds"`
}

func NewCommandSettings(
	logger *zap.Logger,
	k *koanf.Koanf,
	inferenceLoopSettings *reasoning_settings.InferenceLoopSettings,
) *CommandSettings {
	settings := CommandSettings{
		DefaultTimeoutSeconds:    120,
		MaxTimeoutSeconds:        240,
		DefaultYieldSeconds:      30,
		MaxOutputChars:           30_000,
		MaxBackgroundProcesses:   8,
		MaxBackgroundSeconds:     1800,
		FinishedRetentionSeconds: 600,
	}
	if err := k.Unmarshal(CommandSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal command settings", zap.Error(err))
	}

	if settings.DefaultTimeoutSeconds <= 0 {
		logger.Warn("invalid command.default_timeout_seconds, defaulting to 120")
		settings.DefaultTimeoutSeconds = 120
	}
	if settings.MaxTimeoutSeconds <= 0 {
		logger.Warn("invalid command.max_timeout_seconds, defaulting to 240")
		settings.MaxTimeoutSeconds = 240
	}
	// Zero is meaningful here: it turns auto-backgrounding off, so only a negative value is a mistake.
	if settings.DefaultYieldSeconds < 0 {
		logger.Warn("invalid command.default_yield_seconds, defaulting to 30")
		settings.DefaultYieldSeconds = 30
	}
	if settings.MaxOutputChars <= 0 {
		logger.Warn("invalid command.max_output_chars, defaulting to 30000")
		settings.MaxOutputChars = 30_000
	}
	if settings.MaxBackgroundProcesses <= 0 {
		logger.Warn("invalid command.max_background_processes, defaulting to 8")
		settings.MaxBackgroundProcesses = 8
	}
	if settings.MaxBackgroundSeconds <= 0 {
		logger.Warn("invalid command.max_background_seconds, defaulting to 1800")
		settings.MaxBackgroundSeconds = 1800
	}
	if settings.FinishedRetentionSeconds <= 0 {
		logger.Warn("invalid command.finished_retention_seconds, defaulting to 600")
		settings.FinishedRetentionSeconds = 600
	}

	if settings.MaxTimeoutSeconds < settings.DefaultTimeoutSeconds {
		logger.Warn("command.max_timeout_seconds is below command.default_timeout_seconds, raising it to match")
		settings.MaxTimeoutSeconds = settings.DefaultTimeoutSeconds
	}
	// Not clamped: the operator may have raised the loop budget for a reason, but a silent turn loss is worse.
	if settings.MaxTimeoutSeconds >= inferenceLoopSettings.MaxDurationSeconds {
		logger.Warn(
			"command.max_timeout_seconds is not below inference_loop.max_duration_seconds, so a command reaching it will abort the whole turn",
			zap.Int("command.max_timeout_seconds", settings.MaxTimeoutSeconds),
			zap.Int("inference_loop.max_duration_seconds", inferenceLoopSettings.MaxDurationSeconds),
		)
	}

	return &settings
}
