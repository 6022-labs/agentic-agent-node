package services

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-jwt/jwt/v5"
)

type JwtValidator struct {
	settings *settings.JwtSettings
	now      func() time.Time
}

func NewJwtValidator(settings *settings.JwtSettings) *JwtValidator {
	return &JwtValidator{settings: settings, now: time.Now}
}

func (v *JwtValidator) Verify(token string) (common.Address, error) {
	parsed, err := jwt.Parse(
		token,
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
	return common.HexToAddress(sub), nil
}
