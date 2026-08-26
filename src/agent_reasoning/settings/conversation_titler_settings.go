package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const ConversationTitlerSettingsKey = "conversation_titler"

// ConversationTitlerSettings bounds the background labelling round-trip; it runs
// after the turn is answered, so it gets its own budget.
type ConversationTitlerSettings struct {
	TimeoutSeconds int `koanf:"timeout_seconds"`
}

func NewConversationTitlerSettings(logger *zap.Logger, k *koanf.Koanf) *ConversationTitlerSettings {
	settings := ConversationTitlerSettings{
		TimeoutSeconds: 30,
	}
	if err := k.Unmarshal(ConversationTitlerSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal conversation titler settings", zap.Error(err))
	}

	if settings.TimeoutSeconds <= 0 {
		logger.Warn("invalid conversation_titler.timeout_seconds, defaulting to 30")
		settings.TimeoutSeconds = 30
	}

	return &settings
}
