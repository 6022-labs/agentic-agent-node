package settings

import (
	"os"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const ImageGenerationSettingsKey = "image_generation"

// ImageGenerationSettings configures the agent frame; ImageMarkPath is an optional watermark PNG (nil disables it).
type ImageGenerationSettings struct {
	ImageMarkPath *string `koanf:"image_mark_path"`
}

func NewImageGenerationSettings(logger *zap.Logger, k *koanf.Koanf) *ImageGenerationSettings {
	var s ImageGenerationSettings
	if err := k.Unmarshal(ImageGenerationSettingsKey, &s); err != nil {
		logger.Fatal("failed to unmarshal image generation settings", zap.Error(err))
	}

	if s.ImageMarkPath != nil {
		value := strings.TrimSpace(*s.ImageMarkPath)
		if len(value) == 0 {
			s.ImageMarkPath = nil
		} else if _, err := os.Stat(value); os.IsNotExist(err) {
			logger.Warn("image_generation.image_mark_path does not exist", zap.String("path", value))
			s.ImageMarkPath = nil
		} else {
			s.ImageMarkPath = &value
		}
	}

	if s.ImageMarkPath == nil {
		logger.Info("image_generation.image_mark_path is not set; agent images will not be marked")
	}

	return &s
}
