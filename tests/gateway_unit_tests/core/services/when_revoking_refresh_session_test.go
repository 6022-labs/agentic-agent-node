package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/stretchr/testify/assert"
)

type WhenRevokingRefreshSessionTestingSuite struct {
	sut *services.RefreshSessionStore
}

func WhenRevokingRefreshSessionBeforeEach() *WhenRevokingRefreshSessionTestingSuite {
	return &WhenRevokingRefreshSessionTestingSuite{sut: services.NewRefreshSessionStore()}
}

func TestWhenRevokingRefreshSession(t *testing.T) {
	t.Parallel()

	t.Run("Given a live session", func(t *testing.T) {
		t.Parallel()

		t.Run("Should mark it revoked", func(t *testing.T) {
			t.Parallel()

			suite := WhenRevokingRefreshSessionBeforeEach()

			suite.sut.Revoke("session-1", time.Now().Add(time.Hour))

			assert.True(t, suite.sut.IsRevoked("session-1"))
		})

		t.Run("Should leave other sessions untouched", func(t *testing.T) {
			t.Parallel()

			suite := WhenRevokingRefreshSessionBeforeEach()

			suite.sut.Revoke("session-1", time.Now().Add(time.Hour))

			assert.False(t, suite.sut.IsRevoked("session-2"))
		})
	})
}
