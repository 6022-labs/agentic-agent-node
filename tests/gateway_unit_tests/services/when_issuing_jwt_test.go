package services_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestWhenIssuingJwt(t *testing.T) {
	t.Parallel()

	t.Run("Given valid JWT settings and an address", func(t *testing.T) {
		t.Parallel()

		jwtSettings := &settings.JwtSettings{
			SigningKey: "test-signing-key",
			TokenTtl:   30 * time.Minute,
		}
		issuer := services.NewJwtIssuer(jwtSettings)
		address := common.HexToAddress("0x1111111111111111111111111111111111111111")

		token, expiresAt, err := issuer.Issue(address)

		t.Run("Should return a non-empty token", func(t *testing.T) {
			assert.NoError(t, err)
			assert.NotEmpty(t, token)
		})

		t.Run("Should set expiry equal to now + token_ttl", func(t *testing.T) {
			assert.WithinDuration(t, time.Now().Add(30*time.Minute), expiresAt, 5*time.Second)
		})

		t.Run("Should produce a 3-part JWT with HS256 alg and sub == address", func(t *testing.T) {
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
			assert.NotZero(t, claims["iat"])
			assert.NotZero(t, claims["nbf"])
			assert.NotZero(t, claims["exp"])
		})
	})
}
