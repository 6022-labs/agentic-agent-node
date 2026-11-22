package services_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type WhenIssuingJwtTestingSuite struct {
	sut         *services.JwtIssuer
	jwtSettings *settings.JwtSettings
}

func WhenIssuingJwtBeforeEach() *WhenIssuingJwtTestingSuite {
	jwtSettings := &settings.JwtSettings{
		SigningKey: "test-signing-key",
		TokenTtl:   1800 * time.Second,
	}
	return &WhenIssuingJwtTestingSuite{
		sut:         services.NewJwtIssuer(jwtSettings),
		jwtSettings: jwtSettings,
	}
}

func TestWhenIssuingJwt(t *testing.T) {
	t.Parallel()

	t.Run("Given valid JWT settings and an address", func(t *testing.T) {
		t.Parallel()

		address := common.HexToAddress("0x1111111111111111111111111111111111111111")
		sessionExpiresAt := time.Now().Add(7 * 24 * time.Hour)

		t.Run("Should return a non-empty token", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingJwtBeforeEach()

			token, _, err := suite.sut.Issue(address, "session-1", sessionExpiresAt)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})

		t.Run("Should set expiry equal to now + token_ttl", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingJwtBeforeEach()

			_, expiresAt, err := suite.sut.Issue(address, "session-1", sessionExpiresAt)

			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now().Add(suite.jwtSettings.TokenTtl), expiresAt, 5*time.Second)
		})

		t.Run("Should produce a 3-part JWT with HS256 alg, sub == address and sid == session", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingJwtBeforeEach()

			token, _, err := suite.sut.Issue(address, "session-1", sessionExpiresAt)
			assert.NoError(t, err)

			parts := strings.Split(token, ".")
			assert.Len(t, parts, 3)

			headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
			assert.NoError(t, err)
			var header map[string]any
			_ = json.Unmarshal(headerBytes, &header)
			assert.Equal(t, "HS256", header["alg"])
			assert.Equal(t, "JWT", header["typ"])

			claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			assert.NoError(t, err)
			var claims map[string]any
			_ = json.Unmarshal(claimsBytes, &claims)
			assert.Equal(t, address.Hex(), claims["sub"])
			assert.Equal(t, "session-1", claims["sid"])
			assert.NotZero(t, claims["iat"])
			assert.NotZero(t, claims["nbf"])
			assert.NotZero(t, claims["exp"])
		})
	})

	t.Run("Given a session ending before token_ttl", func(t *testing.T) {
		t.Parallel()

		address := common.HexToAddress("0x1111111111111111111111111111111111111111")
		sessionExpiresAt := time.Unix(time.Now().Add(5*time.Minute).Unix(), 0)

		t.Run("Should expire with the session", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingJwtBeforeEach()

			token, expiresAt, err := suite.sut.Issue(address, "session-1", sessionExpiresAt)
			assert.NoError(t, err)

			claimsBytes, err := base64.RawURLEncoding.DecodeString(strings.Split(token, ".")[1])
			assert.NoError(t, err)
			var claims map[string]any
			_ = json.Unmarshal(claimsBytes, &claims)
			assert.Equal(t, sessionExpiresAt, expiresAt)
			assert.Equal(t, float64(sessionExpiresAt.Unix()), claims["exp"])
		})
	})
}
