package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/stretchr/testify/assert"
)

type WhenReissuingRefreshTokenTestingSuite struct {
	sut         *services.RefreshTokenManager
	jwtSettings *settings.JwtSettings
}

func WhenReissuingRefreshTokenBeforeEach() *WhenReissuingRefreshTokenTestingSuite {
	jwtSettings := &settings.JwtSettings{
		SigningKey:     "test-signing-key",
		RefreshTtl:     7 * 24 * time.Hour,
		RefreshIdleTtl: time.Hour,
	}
	return &WhenReissuingRefreshTokenTestingSuite{
		sut:         services.NewRefreshTokenManager(jwtSettings),
		jwtSettings: jwtSettings,
	}
}

func TestWhenReissuingRefreshToken(t *testing.T) {
	t.Parallel()

	address := common_tests.RandomEthAddress()

	t.Run("Given a session ending before the idle expiry", func(t *testing.T) {
		t.Parallel()

		claims := &interfaces.RefreshTokenClaims{
			Address:          address,
			SessionId:        "session-1",
			TokenId:          "token-1",
			SessionExpiresAt: time.Unix(time.Now().Add(10*time.Minute).Unix(), 0),
		}

		t.Run("Should expire with the session", func(t *testing.T) {
			t.Parallel()

			suite := WhenReissuingRefreshTokenBeforeEach()

			reissued, err := suite.sut.Reissue(claims, "next-token-id")

			assert.NoError(t, err)
			assert.Equal(t, claims.SessionExpiresAt, reissued.ExpiresAt)
		})
	})

	t.Run("Given the claims of an issued token and a new token id", func(t *testing.T) {
		t.Parallel()

		t.Run("Should keep the session", func(t *testing.T) {
			t.Parallel()

			suite := WhenReissuingRefreshTokenBeforeEach()
			issued, _ := suite.sut.Issue(address)
			claims, _ := suite.sut.Parse(issued.Token)

			reissued, err := suite.sut.Reissue(claims, "next-token-id")
			assert.NoError(t, err)
			reissuedClaims, err := suite.sut.Parse(reissued.Token)

			assert.NoError(t, err)
			assert.Equal(t, claims.SessionId, reissuedClaims.SessionId)
			assert.Equal(t, claims.SessionExpiresAt, reissuedClaims.SessionExpiresAt)
		})

		t.Run("Should expire refresh_idle_ttl from now", func(t *testing.T) {
			t.Parallel()

			suite := WhenReissuingRefreshTokenBeforeEach()
			issued, _ := suite.sut.Issue(address)
			claims, _ := suite.sut.Parse(issued.Token)

			reissued, err := suite.sut.Reissue(claims, "next-token-id")

			assert.NoError(t, err)
			assert.WithinDuration(
				t,
				time.Now().Add(suite.jwtSettings.RefreshIdleTtl),
				reissued.ExpiresAt,
				5*time.Second,
			)
		})

		t.Run("Should carry the new token id", func(t *testing.T) {
			t.Parallel()

			suite := WhenReissuingRefreshTokenBeforeEach()
			issued, _ := suite.sut.Issue(address)
			claims, _ := suite.sut.Parse(issued.Token)

			reissued, _ := suite.sut.Reissue(claims, "next-token-id")
			reissuedClaims, err := suite.sut.Parse(reissued.Token)

			assert.NoError(t, err)
			assert.Equal(t, "next-token-id", reissuedClaims.TokenId)
		})
	})
}
