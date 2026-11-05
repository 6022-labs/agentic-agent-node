package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

type WhenIssuingRefreshTokenTestingSuite struct {
	sut         *services.RefreshTokenManager
	jwtSettings *settings.JwtSettings
}

func WhenIssuingRefreshTokenBeforeEach(jwtSettings *settings.JwtSettings) *WhenIssuingRefreshTokenTestingSuite {
	return &WhenIssuingRefreshTokenTestingSuite{
		sut:         services.NewRefreshTokenManager(jwtSettings),
		jwtSettings: jwtSettings,
	}
}

func TestWhenIssuingRefreshToken(t *testing.T) {
	t.Parallel()

	address := common_tests.RandomEthAddress()

	t.Run("Given a refresh_ttl shorter than refresh_idle_ttl", func(t *testing.T) {
		t.Parallel()

		jwtSettings := &settings.JwtSettings{
			SigningKey:     "test-signing-key",
			RefreshTtl:     time.Hour,
			RefreshIdleTtl: 24 * time.Hour,
		}

		t.Run("Should expire with the session", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingRefreshTokenBeforeEach(jwtSettings)

			refreshToken, err := suite.sut.Issue(address)

			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now().Add(jwtSettings.RefreshTtl), refreshToken.ExpiresAt, 5*time.Second)
		})
	})

	t.Run("Given valid JWT settings and an address", func(t *testing.T) {
		t.Parallel()

		jwtSettings := &settings.JwtSettings{
			SigningKey:     "test-signing-key",
			RefreshTtl:     7 * 24 * time.Hour,
			RefreshIdleTtl: 24 * time.Hour,
		}

		t.Run("Should return a token that parses back to the address", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingRefreshTokenBeforeEach(jwtSettings)

			refreshToken, err := suite.sut.Issue(address)
			assert.NoError(t, err)
			claims, err := suite.sut.Parse(refreshToken.Token)

			assert.NoError(t, err)
			assert.Equal(t, address, claims.Address)
		})

		t.Run("Should expire refresh_idle_ttl from now", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingRefreshTokenBeforeEach(jwtSettings)

			refreshToken, err := suite.sut.Issue(address)

			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now().Add(jwtSettings.RefreshIdleTtl), refreshToken.ExpiresAt, 5*time.Second)
		})

		t.Run("Should end the session refresh_ttl from now", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingRefreshTokenBeforeEach(jwtSettings)

			refreshToken, _ := suite.sut.Issue(address)
			claims, err := suite.sut.Parse(refreshToken.Token)

			assert.NoError(t, err)
			assert.WithinDuration(t, time.Now().Add(jwtSettings.RefreshTtl), claims.SessionExpiresAt, 5*time.Second)
		})

		t.Run("Should start a new session on every call", func(t *testing.T) {
			t.Parallel()

			suite := WhenIssuingRefreshTokenBeforeEach(jwtSettings)

			first, _ := suite.sut.Issue(address)
			second, _ := suite.sut.Issue(address)
			firstClaims, _ := suite.sut.Parse(first.Token)
			secondClaims, _ := suite.sut.Parse(second.Token)

			assert.NotEqual(t, firstClaims.SessionId, secondClaims.SessionId)
		})
	})
}
