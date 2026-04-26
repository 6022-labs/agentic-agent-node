package settings

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const JwtSettingsKey = "jwt"

type JwtSettings struct {
	SigningKey string        `koanf:"signing_key"`
	TokenTtl   time.Duration `koanf:"token_ttl"`
	NonceTtl   time.Duration `koanf:"nonce_ttl"`
}

func NewJwtSettings(logger *zap.Logger, k *koanf.Koanf) *JwtSettings {
	settings := JwtSettings{
		TokenTtl: 30 * time.Minute,
		NonceTtl: 5 * time.Minute,
	}
	if err := k.Unmarshal(JwtSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal jwt settings", zap.Error(err))
	}

	settings.SigningKey = strings.TrimSpace(settings.SigningKey)

	if len(settings.SigningKey) == 0 {
		logger.Fatal("please set jwt.signing_key (or JWT__SIGNING_KEY env)")
	}
	if settings.TokenTtl <= 0 {
		logger.Warn("invalid jwt.token_ttl value, defaulting to 30m")
		settings.TokenTtl = 30 * time.Minute
	}
	if settings.NonceTtl <= 0 {
		logger.Warn("invalid jwt.nonce_ttl value, defaulting to 5m")
		settings.NonceTtl = 5 * time.Minute
	}

	return &settings
}
