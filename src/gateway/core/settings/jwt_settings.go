package settings

import (
	"strings"
	"time"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const JwtSettingsKey = "jwt"

const minSigningKeyLength = 32

type JwtSettings struct {
	SigningKey     string
	TokenTtl       time.Duration
	NonceTtl       time.Duration
	RefreshTtl     time.Duration
	RefreshIdleTtl time.Duration
}

type jwtSettingsRaw struct {
	SigningKey            string `koanf:"signing_key"`
	TokenTtlSeconds       int    `koanf:"token_ttl"`
	NonceTtlSeconds       int    `koanf:"nonce_ttl"`
	RefreshTtlSeconds     int    `koanf:"refresh_ttl"`
	RefreshIdleTtlSeconds int    `koanf:"refresh_idle_ttl"`
}

func NewJwtSettings(logger *zap.Logger, k *koanf.Koanf) *JwtSettings {
	raw := jwtSettingsRaw{
		TokenTtlSeconds:       1800,
		NonceTtlSeconds:       300,
		RefreshTtlSeconds:     604800,
		RefreshIdleTtlSeconds: 86400,
	}
	if err := k.Unmarshal(JwtSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal jwt settings", zap.Error(err))
	}

	raw.SigningKey = strings.TrimSpace(raw.SigningKey)

	if len(raw.SigningKey) < minSigningKeyLength {
		logger.Fatal("jwt.signing_key must be at least 32 characters (or JWT__SIGNING_KEY env)")
	}
	if raw.TokenTtlSeconds <= 0 {
		logger.Warn("invalid jwt.token_ttl value, defaulting to 1800")
		raw.TokenTtlSeconds = 1800
	}
	if raw.NonceTtlSeconds <= 0 {
		logger.Warn("invalid jwt.nonce_ttl value, defaulting to 300")
		raw.NonceTtlSeconds = 300
	}
	if raw.RefreshTtlSeconds <= 0 {
		logger.Warn("invalid jwt.refresh_ttl value, defaulting to 604800")
		raw.RefreshTtlSeconds = 604800
	}
	if raw.RefreshIdleTtlSeconds <= 0 {
		logger.Warn("invalid jwt.refresh_idle_ttl value, defaulting to 86400")
		raw.RefreshIdleTtlSeconds = 86400
	}
	if raw.RefreshIdleTtlSeconds < 2*raw.TokenTtlSeconds {
		logger.Warn("jwt.refresh_idle_ttl is under twice jwt.token_ttl, raising it")
		raw.RefreshIdleTtlSeconds = 2 * raw.TokenTtlSeconds
	}

	return &JwtSettings{
		SigningKey:     raw.SigningKey,
		TokenTtl:       time.Duration(raw.TokenTtlSeconds) * time.Second,
		NonceTtl:       time.Duration(raw.NonceTtlSeconds) * time.Second,
		RefreshTtl:     time.Duration(raw.RefreshTtlSeconds) * time.Second,
		RefreshIdleTtl: time.Duration(raw.RefreshIdleTtlSeconds) * time.Second,
	}
}
