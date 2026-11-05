package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/ethereum/go-ethereum/common"
	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type WhenValidatingJwtTestingSuite struct {
	jwtSettings *settings.JwtSettings
	sut         *services.JwtValidator
}

func WhenValidatingJwtBeforeEach() *WhenValidatingJwtTestingSuite {
	jwtSettings := &settings.JwtSettings{
		SigningKey:     "test-signing-key-at-least-32-characters",
		TokenTtl:       time.Hour,
		RefreshTtl:     7 * 24 * time.Hour,
		RefreshIdleTtl: 24 * time.Hour,
	}

	return &WhenValidatingJwtTestingSuite{
		jwtSettings: jwtSettings,
		sut:         services.NewJwtValidator(jwtSettings),
	}
}

func TestWhenValidatingJwt(t *testing.T) {
	t.Parallel()

	sign := func(method gojwt.SigningMethod, claims gojwt.MapClaims) string {
		signed, _ := gojwt.NewWithClaims(method, claims).SignedString([]byte("test-signing-key-at-least-32-characters"))
		return signed
	}
	validClaims := func() gojwt.MapClaims {
		return gojwt.MapClaims{
			"sub": common_tests.RandomEthAddress().Hex(),
			"exp": time.Now().Add(time.Hour).Unix(),
		}
	}

	t.Run("Given a malformed token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()

			result, err := suite.sut.Verify("not-a-jwt")

			assert.ErrorIs(t, err, gojwt.ErrTokenMalformed)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a token signed with another key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			other := services.NewJwtIssuer(
				&settings.JwtSettings{SigningKey: "another-signing-key-at-least-32-chars", TokenTtl: time.Hour},
			)
			token, _, _ := other.Issue(common_tests.RandomEthAddress())

			result, err := suite.sut.Verify(token)

			assert.ErrorIs(t, err, gojwt.ErrTokenSignatureInvalid)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a refresh token signed with the same key", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			refreshToken, _ := services.NewRefreshTokenManager(suite.jwtSettings).Issue(common_tests.RandomEthAddress())
			token := refreshToken.Token

			result, err := suite.sut.Verify(token)

			assert.ErrorIs(t, err, gojwt.ErrTokenSignatureInvalid)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a token signed with another algorithm", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()

			result, err := suite.sut.Verify(sign(gojwt.SigningMethodHS512, validClaims()))

			assert.ErrorIs(t, err, gojwt.ErrTokenSignatureInvalid)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given an expired token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			claims := validClaims()
			claims["exp"] = time.Now().Add(-time.Minute).Unix()

			result, err := suite.sut.Verify(sign(gojwt.SigningMethodHS256, claims))

			assert.ErrorIs(t, err, gojwt.ErrTokenExpired)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a token without an expiry", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			claims := validClaims()
			delete(claims, "exp")

			result, err := suite.sut.Verify(sign(gojwt.SigningMethodHS256, claims))

			assert.ErrorIs(t, err, gojwt.ErrTokenRequiredClaimMissing)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a token without a subject", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			claims := validClaims()
			delete(claims, "sub")

			result, err := suite.sut.Verify(sign(gojwt.SigningMethodHS256, claims))

			assert.ErrorIs(t, err, gojwt.ErrTokenInvalidSubject)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a token with a non-address subject", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			claims := validClaims()
			claims["sub"] = "not-an-address"

			result, err := suite.sut.Verify(sign(gojwt.SigningMethodHS256, claims))

			assert.ErrorIs(t, err, gojwt.ErrTokenInvalidSubject)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a token with a non-string subject", func(t *testing.T) {
		t.Parallel()

		t.Run("Should reject it", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			claims := validClaims()
			claims["sub"] = 42

			result, err := suite.sut.Verify(sign(gojwt.SigningMethodHS256, claims))

			assert.ErrorIs(t, err, gojwt.ErrTokenInvalidSubject)
			assert.Equal(t, common.Address{}, result)
		})
	})

	t.Run("Given a token it issued", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return the address", func(t *testing.T) {
			t.Parallel()

			suite := WhenValidatingJwtBeforeEach()
			address := common_tests.RandomEthAddress()
			token, _, _ := services.NewJwtIssuer(suite.jwtSettings).Issue(address)

			result, err := suite.sut.Verify(token)

			assert.NoError(t, err)
			assert.Equal(t, address, result)
		})
	})
}
