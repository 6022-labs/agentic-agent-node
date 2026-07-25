package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const InferenceLoopSettingsKey = "inference_loop"

// InferenceLoopSettings bounds the LLM ↔ MCP loop: round-trips (MaxIterations),
// wall-clock (MaxDurationSeconds), and total LLM tokens across iterations (MaxTotalTokens).
type InferenceLoopSettings struct {
	MaxIterations      int `koanf:"max_iterations"`
	MaxDurationSeconds int `koanf:"max_duration_seconds"`
	MaxTotalTokens     int `koanf:"max_total_tokens"`
}

func NewInferenceLoopSettings(logger *zap.Logger, k *koanf.Koanf) *InferenceLoopSettings {
	settings := InferenceLoopSettings{
		MaxIterations:      10,
		MaxDurationSeconds: 300,
		MaxTotalTokens:     500_000,
	}
	if err := k.Unmarshal(InferenceLoopSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal inner loop settings", zap.Error(err))
	}

	if settings.MaxIterations <= 0 {
		logger.Warn("invalid inference_loop.max_iterations, defaulting to 10")
		settings.MaxIterations = 10
	}
	if settings.MaxDurationSeconds <= 0 {
		logger.Warn("invalid inference_loop.max_duration_seconds, defaulting to 300")
		settings.MaxDurationSeconds = 300
	}
	if settings.MaxTotalTokens <= 0 {
		logger.Warn("invalid inference_loop.max_total_tokens, defaulting to 500000")
		settings.MaxTotalTokens = 500_000
	}

	return &settings
}
