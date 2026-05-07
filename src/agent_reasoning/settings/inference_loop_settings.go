package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const InferenceLoopSettingsKey = "inference_loop"

// InferenceLoopSettings bounds the LLM ↔ MCP loop. MaxIterations is the hard cap
// on how many LLM round-trips a single InferenceLoop.Run can make before giving up
// — stops a runaway agent from burning tokens forever.
type InferenceLoopSettings struct {
	MaxIterations int `koanf:"max_iterations"`
}

func NewInferenceLoopSettings(logger *zap.Logger, k *koanf.Koanf) *InferenceLoopSettings {
	settings := InferenceLoopSettings{
		MaxIterations: 10,
	}
	if err := k.Unmarshal(InferenceLoopSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal inner loop settings", zap.Error(err))
	}

	if settings.MaxIterations <= 0 {
		logger.Warn("invalid inference_loop.max_iterations, defaulting to 10")
		settings.MaxIterations = 10
	}

	return &settings
}
