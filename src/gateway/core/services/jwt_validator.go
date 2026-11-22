package services

import (
	"errors"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrAccessTokenSessionMissing = errors.New("access token has no session")
	ErrAccessTokenSessionRevoked = errors.New("access token session revoked")
)

type JwtValidator struct {
	settings            *settings.JwtSettings
	refreshSessionStore interfaces.RefreshSessionStoreInterface
	now                 func() time.Time
}

func NewJwtValidator(
	settings *settings.JwtSettings,
	refreshSessionStore interfaces.RefreshSessionStoreInterface,
) *JwtValidator {
	return &JwtValidator{settings: settings, refreshSessionStore: refreshSessionStore, now: time.Now}
}

func (v *JwtValidator) Verify(token string) (common.Address, error) {
	claims := jwt.MapClaims{}
	parsed, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			return []byte(v.settings.SigningKey), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return v.now() }),
	)
	if err != nil {
		return common.Address{}, err
	}

	sub, err := parsed.Claims.GetSubject()
	if err != nil || !common.IsHexAddress(sub) {
		return common.Address{}, jwt.ErrTokenInvalidSubject
	}

	sessionId, _ := claims[sessionIdClaim].(string)
	if sessionId == "" {
		return common.Address{}, ErrAccessTokenSessionMissing
	}
	if v.refreshSessionStore.IsRevoked(sessionId) {
		return common.Address{}, ErrAccessTokenSessionRevoked
	}
	return common.HexToAddress(sub), nil
}
