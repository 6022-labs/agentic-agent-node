package settings

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const JwtSettingsKey = "jwt"

type JwtSettings struct {
	SigningKey string
	TokenTtl   time.Duration
	NonceTtl   time.Duration
}

type jwtSettingsRaw struct {
	SigningKey      string `koanf:"signing_key"`
	TokenTtlSeconds int    `koanf:"token_ttl"`
	NonceTtlSeconds int    `koanf:"nonce_ttl"`
}

func NewJwtSettings(logger *zap.Logger, k *koanf.Koanf) *JwtSettings {
	raw := jwtSettingsRaw{
		TokenTtlSeconds: 1800,
		NonceTtlSeconds: 300,
	}
	if err := k.Unmarshal(JwtSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal jwt settings", zap.Error(err))
	}

	raw.SigningKey = strings.TrimSpace(raw.SigningKey)

	if len(raw.SigningKey) == 0 {
		logger.Fatal("please set jwt.signing_key (or JWT__SIGNING_KEY env)")
	}
	if raw.TokenTtlSeconds <= 0 {
		logger.Warn("invalid jwt.token_ttl value, defaulting to 1800")
		raw.TokenTtlSeconds = 1800
	}
	if raw.NonceTtlSeconds <= 0 {
		logger.Warn("invalid jwt.nonce_ttl value, defaulting to 300")
		raw.NonceTtlSeconds = 300
	}

	return &JwtSettings{
		SigningKey: raw.SigningKey,
		TokenTtl:   time.Duration(raw.TokenTtlSeconds) * time.Second,
		NonceTtl:   time.Duration(raw.NonceTtlSeconds) * time.Second,
	}
}
