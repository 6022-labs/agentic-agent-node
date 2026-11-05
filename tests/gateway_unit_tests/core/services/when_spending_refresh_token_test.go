package services_test

import (
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services/interfaces"
	"github.com/stretchr/testify/assert"
)

type WhenSpendingRefreshTokenTestingSuite struct {
	sut *services.RefreshSessionStore
}

func WhenSpendingRefreshTokenBeforeEach() *WhenSpendingRefreshTokenTestingSuite {
	return &WhenSpendingRefreshTokenTestingSuite{sut: services.NewRefreshSessionStore()}
}

func TestWhenSpendingRefreshToken(t *testing.T) {
	t.Parallel()

	spent := interfaces.SpentRefreshToken{SuccessorId: "successor-1", SpentAt: time.Now()}

	t.Run("Given a spent token whose session has expired", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSpendingRefreshTokenTestingSuite) {
			suite.sut.Spend("session-1", "token-1", spent, time.Now().Add(-time.Second))
		}

		t.Run("Should forget it", func(t *testing.T) {
			t.Parallel()

			suite := WhenSpendingRefreshTokenBeforeEach()
			initSuite(suite)

			_, spentNow := suite.sut.Spend("session-1", "token-1", spent, time.Now().Add(time.Hour))

			assert.True(t, spentNow)
		})
	})
	t.Run("Given an already spent token", func(t *testing.T) {
		t.Parallel()

		initSuite := func(suite *WhenSpendingRefreshTokenTestingSuite) {
			suite.sut.Spend("session-1", "token-1", spent, time.Now().Add(time.Hour))
		}

		t.Run("Should return the earlier record", func(t *testing.T) {
			t.Parallel()

			suite := WhenSpendingRefreshTokenBeforeEach()
			initSuite(suite)
			again := interfaces.SpentRefreshToken{SuccessorId: "successor-2", SpentAt: time.Now()}

			earlier, spentNow := suite.sut.Spend("session-1", "token-1", again, time.Now().Add(time.Hour))

			assert.False(t, spentNow)
			assert.Equal(t, spent, *earlier)
		})
	})

	t.Run("Given an unspent token", func(t *testing.T) {
		t.Parallel()

		t.Run("Should spend it", func(t *testing.T) {
			t.Parallel()

			suite := WhenSpendingRefreshTokenBeforeEach()

			earlier, spentNow := suite.sut.Spend("session-1", "token-1", spent, time.Now().Add(time.Hour))

			assert.True(t, spentNow)
			assert.Nil(t, earlier)
		})
	})
}
