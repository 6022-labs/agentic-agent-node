package services_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

type WhenValidatingJwtTestingSuite struct {
	sut         *services.JwtValidator
	issuer      *services.JwtIssuer
	jwtSettings *settings.JwtSettings
}

func WhenValidatingJwtBeforeEach() *WhenValidatingJwtTestingSuite {
	jwtSettings := &settings.JwtSettings{
		SigningKey: "round-trip-key",
		TokenTtl:   1800 * time.Second,
	}
	return &WhenValidatingJwtTestingSuite{
		sut:         services.NewJwtValidator(jwtSettings),
		issuer:      services.NewJwtIssuer(jwtSettings),
		jwtSettings: jwtSettings,
	}
}

func TestWhenValidatingJwt(t *testing.T) {
	t.Parallel()

	address := common.HexToAddress("0x2222222222222222222222222222222222222222")

	t.Run("Given a malformed token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject input without three segments", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()

			_, err := suite.sut.Verify("not.a.token.really")
			assert.ErrorContains(t, err, "invalid number of segments")
		})

		t.Run("Should reject a non-base64 signature", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()

			_, err := suite.sut.Verify("aGVhZGVy.cGF5bG9hZA.??not-base64??")
			assert.ErrorContains(t, err, "token is malformed")
		})
	})

	t.Run("Given a token signed with a different key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject as signature mismatch", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()

			otherIssuer := services.NewJwtIssuer(&settings.JwtSettings{
				SigningKey: "other-key",
				TokenTtl:   1800 * time.Second,
			})
			token, _, _ := otherIssuer.Issue(address)

			_, err := suite.sut.Verify(token)

			assert.ErrorContains(t, err, "signature is invalid")
		})
	})

	t.Run("Given a token whose exp is in the past", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject as expired", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()

			token := craftToken(
				t,
				suite.jwtSettings.SigningKey,
				address.Hex(),
				time.Now().Add(-time.Hour).Unix(),
				time.Now().Add(-2*time.Hour).Unix(),
			)

			_, err := suite.sut.Verify(token)

			assert.ErrorContains(t, err, "expired")
		})
	})

	t.Run("Given a token issued with the same settings", func(t *testing.T) {
		t.Parallel()

		t.Run("Should recover the original address", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()

			token, _, _ := suite.issuer.Issue(address)

			recovered, err := suite.sut.Verify(token)

			assert.NoError(t, err)
			assert.Equal(t, address, recovered)
		})
	})
}

func craftToken(t *testing.T, key, sub string, exp, nbf int64) string {
	t.Helper()
	encoder := base64.RawURLEncoding
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{"sub": sub, "iat": time.Now().Unix(), "nbf": nbf, "exp": exp}
	headerJson, _ := json.Marshal(header)
	claimsJson, _ := json.Marshal(claims)
	signing := encoder.EncodeToString(headerJson) + "." + encoder.EncodeToString(claimsJson)
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(signing))
	return signing + "." + encoder.EncodeToString(mac.Sum(nil))
}
