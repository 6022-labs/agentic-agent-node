package settings

import (
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const SiweSettingsKey = "siwe"

// SiweSettings are the server-controlled fields of the EIP-4361 sign-in message.
type SiweSettings struct {
	Statement string `koanf:"statement"`
	// Uri and Domain come from console.url; Domain is the authority wallets show.
	Uri    string `koanf:"-"`
	Domain string `koanf:"-"`
}

func NewSiweSettings(logger *zap.Logger, k *koanf.Koanf, consoleSettings *ConsoleSettings) *SiweSettings {
	settings := SiweSettings{Statement: "Sign in to your 6022 agent node."}
	if err := k.Unmarshal(SiweSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal siwe settings", zap.Error(err))
	}

	settings.Uri = consoleSettings.Url
	settings.Domain = consoleSettings.Domain

	return &settings
}
