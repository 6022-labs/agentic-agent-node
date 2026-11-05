package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/stretchr/testify/assert"
)

type WhenCheckingRefreshSessionRevocationTestingSuite struct {
	sut *services.RefreshSessionStore
}

func WhenCheckingRefreshSessionRevocationBeforeEach() *WhenCheckingRefreshSessionRevocationTestingSuite {
	return &WhenCheckingRefreshSessionRevocationTestingSuite{sut: services.NewRefreshSessionStore()}
}

func TestWhenCheckingRefreshSessionRevocation(t *testing.T) {
	t.Parallel()

	t.Run("Given an unknown session", func(t *testing.T) {
		t.Parallel()

		t.Run("Should report it not revoked", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingRefreshSessionRevocationBeforeEach()

			assert.False(t, suite.sut.IsRevoked("session-1"))
		})
	})

	t.Run("Given a revocation past its session expiry", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingRefreshSessionRevocationTestingSuite) {
			suite.sut.Revoke("session-1", time.Now().Add(-time.Second))
		}

		t.Run("Should report it not revoked", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingRefreshSessionRevocationBeforeEach()
			initSuite(suite)

			assert.False(t, suite.sut.IsRevoked("session-1"))
		})
	})

	t.Run("Given a revoked live session", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenCheckingRefreshSessionRevocationTestingSuite) {
			suite.sut.Revoke("session-1", time.Now().Add(time.Hour))
		}

		t.Run("Should report it revoked", func(t *testing.T) {
			t.Parallel()

			suite := WhenCheckingRefreshSessionRevocationBeforeEach()
			initSuite(suite)

			assert.True(t, suite.sut.IsRevoked("session-1"))
		})
	})
}
