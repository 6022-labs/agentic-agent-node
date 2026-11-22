package services

import (
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-jwt/jwt/v5"
)

type JwtIssuer struct {
	jwtSettings *settings.JwtSettings
	now         func() time.Time
}

func NewJwtIssuer(jwtSettings *settings.JwtSettings) *JwtIssuer {
	return &JwtIssuer{jwtSettings: jwtSettings, now: time.Now}
}

func (i *JwtIssuer) Issue(
	address common.Address,
	sessionId string,
	sessionExpiresAt time.Time,
) (string, time.Time, error) {
	now := i.now()
	expiresAt := now.Add(i.jwtSettings.TokenTtl)
	if sessionExpiresAt.Before(expiresAt) {
		expiresAt = sessionExpiresAt
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		subjectClaim:   address.Hex(),
		issuedAtClaim:  now.Unix(),
		notBeforeClaim: now.Unix(),
		expiresAtClaim: expiresAt.Unix(),
		sessionIdClaim: sessionId,
	})
	signed, err := token.SignedString([]byte(i.jwtSettings.SigningKey))
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}
