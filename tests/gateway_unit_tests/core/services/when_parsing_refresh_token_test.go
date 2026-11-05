package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type WhenParsingRefreshTokenTestingSuite struct {
	sut         *services.RefreshTokenManager
	jwtSettings *settings.JwtSettings
}

func WhenParsingRefreshTokenBeforeEach() *WhenParsingRefreshTokenTestingSuite {
	jwtSettings := &settings.JwtSettings{
		SigningKey:     "test-signing-key",
		TokenTtl:       30 * time.Minute,
		RefreshTtl:     7 * 24 * time.Hour,
		RefreshIdleTtl: 24 * time.Hour,
	}
	return &WhenParsingRefreshTokenTestingSuite{
		sut:         services.NewRefreshTokenManager(jwtSettings),
		jwtSettings: jwtSettings,
	}
}

func (s *WhenParsingRefreshTokenTestingSuite) signWithRefreshKey(claims jwt.MapClaims) string {
	signed, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(s.jwtSettings.SigningKey + ":refresh"))
	return signed
}

func TestWhenParsingRefreshToken(t *testing.T) {
	t.Parallel()

	address := common_tests.RandomEthAddress()

	t.Run("Given a malformed token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return an error", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()

			_, err := suite.sut.Parse("not-a-jwt")

			assert.ErrorIs(t, err, jwt.ErrTokenMalformed)
		})
	})

	t.Run("Given an access token signed with the same signing key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			accessToken, _, _ := services.NewJwtIssuer(suite.jwtSettings).Issue(address)

			_, err := suite.sut.Parse(accessToken)

			assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
		})
	})

	t.Run("Given a refresh token signed with another key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			other := services.NewRefreshTokenManager(&settings.JwtSettings{
				SigningKey: "another-signing-key",
				RefreshTtl: time.Hour,
			})
			refreshToken, _ := other.Issue(address)

			_, err := suite.sut.Parse(refreshToken.Token)

			assert.ErrorIs(t, err, jwt.ErrTokenSignatureInvalid)
		})
	})

	t.Run("Given an expired refresh token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			expired := services.NewRefreshTokenManager(&settings.JwtSettings{
				SigningKey:     suite.jwtSettings.SigningKey,
				RefreshTtl:     -time.Minute,
				RefreshIdleTtl: 24 * time.Hour,
			})
			refreshToken, _ := expired.Issue(address)

			_, err := suite.sut.Parse(refreshToken.Token)

			assert.ErrorIs(t, err, jwt.ErrTokenExpired)
		})
	})

	t.Run("Given a token signed with the refresh key but another type", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrInvalidRefreshToken", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			token := suite.signWithRefreshKey(jwt.MapClaims{
				"sub": address.Hex(),
				"exp": time.Now().Add(time.Hour).Unix(),
				"typ": "access",
				"sid": "session-1",
				"jti": "token-1",
				"sxp": time.Now().Add(time.Hour).Unix(),
			})

			_, err := suite.sut.Parse(token)

			assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
		})
	})

	t.Run("Given a token signed with the refresh key but no session id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrInvalidRefreshToken", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			token := suite.signWithRefreshKey(jwt.MapClaims{
				"sub": address.Hex(), "exp": time.Now().Add(time.Hour).Unix(), "typ": "refresh", "jti": "token-1",
				"sxp": time.Now().Add(time.Hour).Unix(),
			})

			_, err := suite.sut.Parse(token)

			assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
		})
	})

	t.Run("Given a token signed with the refresh key but no token id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrInvalidRefreshToken", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			token := suite.signWithRefreshKey(jwt.MapClaims{
				"sub": address.Hex(), "exp": time.Now().Add(time.Hour).Unix(), "typ": "refresh", "sid": "session-1",
				"sxp": time.Now().Add(time.Hour).Unix(),
			})

			_, err := suite.sut.Parse(token)

			assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
		})
	})

	t.Run("Given a token signed with the refresh key but a non-address subject", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrInvalidRefreshToken", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			token := suite.signWithRefreshKey(jwt.MapClaims{
				"sub": "not-an-address",
				"exp": time.Now().Add(time.Hour).Unix(),
				"typ": "refresh",
				"sid": "session-1",
				"jti": "token-1",
				"sxp": time.Now().Add(time.Hour).Unix(),
			})

			_, err := suite.sut.Parse(token)

			assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
		})
	})

	t.Run("Given a token signed with the refresh key but no session expiry", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return ErrInvalidRefreshToken", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			token := suite.signWithRefreshKey(jwt.MapClaims{
				"sub": address.Hex(),
				"exp": time.Now().Add(time.Hour).Unix(),
				"typ": "refresh",
				"sid": "session-1",
				"jti": "token-1",
			})

			_, err := suite.sut.Parse(token)

			assert.ErrorIs(t, err, services.ErrInvalidRefreshToken)
		})
	})

	t.Run("Given a valid refresh token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return its claims", func(t *testing.T) {
			t.Parallel()

			suite := WhenParsingRefreshTokenBeforeEach()
			refreshToken, _ := suite.sut.Issue(address)

			claims, err := suite.sut.Parse(refreshToken.Token)

			assert.NoError(t, err)
			assert.Equal(t, address, claims.Address)
			assert.NotEmpty(t, claims.SessionId)
			assert.NotEmpty(t, claims.TokenId)
			assert.Equal(t, refreshToken.ExpiresAt, claims.ExpiresAt)
			assert.WithinDuration(
				t,
				time.Now().Add(suite.jwtSettings.RefreshTtl),
				claims.SessionExpiresAt,
				5*time.Second,
			)
		})
	})
}
