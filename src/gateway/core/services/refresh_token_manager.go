package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-jwt/jwt/v5"
)

const (
	refreshTokenType      = "refresh"
	refreshKeySuffix      = ":refresh"
	refreshIdByteLength   = 16
	refreshTokenTypeClaim = "typ"
	refreshSessionIdClaim = "sid"
	refreshTokenIdClaim   = "jti"
	sessionExpiresAtClaim = "sxp"
)

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type RefreshTokenManager struct {
	jwtSettings *settings.JwtSettings
	now         func() time.Time
}

func NewRefreshTokenManager(jwtSettings *settings.JwtSettings) *RefreshTokenManager {
	return &RefreshTokenManager{jwtSettings: jwtSettings, now: time.Now}
}

func (m *RefreshTokenManager) Issue(address common.Address) (*interfaces.RefreshToken, error) {
	sessionId, err := m.NewTokenId()
	if err != nil {
		return nil, err
	}
	tokenId, err := m.NewTokenId()
	if err != nil {
		return nil, err
	}
	return m.sign(address, sessionId, tokenId, m.now().Add(m.jwtSettings.RefreshTtl))
}

func (m *RefreshTokenManager) Reissue(
	claims *interfaces.RefreshTokenClaims,
	tokenId string,
) (*interfaces.RefreshToken, error) {
	return m.sign(claims.Address, claims.SessionId, tokenId, claims.SessionExpiresAt)
}

func (m *RefreshTokenManager) NewTokenId() (string, error) {
	bytes := make([]byte, refreshIdByteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (m *RefreshTokenManager) Parse(token string) (*interfaces.RefreshTokenClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(
		token,
		claims,
		func(t *jwt.Token) (any, error) {
			return m.key(), nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(func() time.Time { return m.now() }),
	)
	if err != nil {
		return nil, err
	}

	sub, _ := claims.GetSubject()
	tokenType, _ := claims[refreshTokenTypeClaim].(string)
	sessionId, _ := claims[refreshSessionIdClaim].(string)
	tokenId, _ := claims[refreshTokenIdClaim].(string)
	sessionExpiresAt, _ := claims[sessionExpiresAtClaim].(float64)
	if tokenType != refreshTokenType || sessionId == "" || tokenId == "" || sessionExpiresAt <= 0 ||
		!common.IsHexAddress(sub) {
		return nil, ErrInvalidRefreshToken
	}

	return &interfaces.RefreshTokenClaims{
		Address:          common.HexToAddress(sub),
		SessionId:        sessionId,
		TokenId:          tokenId,
		SessionExpiresAt: time.Unix(int64(sessionExpiresAt), 0),
	}, nil
}

func (m *RefreshTokenManager) sign(
	address common.Address,
	sessionId string,
	tokenId string,
	sessionExpiresAt time.Time,
) (*interfaces.RefreshToken, error) {
	now := m.now()
	expiresAt := now.Add(m.jwtSettings.RefreshIdleTtl)
	if sessionExpiresAt.Before(expiresAt) {
		expiresAt = sessionExpiresAt
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":                 address.Hex(),
		"iat":                 now.Unix(),
		"nbf":                 now.Unix(),
		"exp":                 expiresAt.Unix(),
		refreshTokenTypeClaim: refreshTokenType,
		refreshSessionIdClaim: sessionId,
		refreshTokenIdClaim:   tokenId,
		sessionExpiresAtClaim: sessionExpiresAt.Unix(),
	})
	signed, err := token.SignedString(m.key())
	if err != nil {
		return nil, err
	}
	return &interfaces.RefreshToken{Token: signed, ExpiresAt: time.Unix(expiresAt.Unix(), 0)}, nil
}

func (m *RefreshTokenManager) key() []byte {
	return []byte(m.jwtSettings.SigningKey + refreshKeySuffix)
}
