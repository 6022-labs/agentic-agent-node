package services_test

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/services"
	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/stretchr/testify/assert"
)

type WhenGeneratingRefreshTokenIdTestingSuite struct {
	sut *services.RefreshTokenManager
}

func WhenGeneratingRefreshTokenIdBeforeEach() *WhenGeneratingRefreshTokenIdTestingSuite {
	return &WhenGeneratingRefreshTokenIdTestingSuite{
		sut: services.NewRefreshTokenManager(&settings.JwtSettings{
			SigningKey: "test-signing-key",
			RefreshTtl: time.Hour,
		}),
	}
}

func TestWhenGeneratingRefreshTokenId(t *testing.T) {
	t.Parallel()

	t.Run("Given a working entropy source", func(t *testing.T) {
		t.Parallel()

		t.Run("Should return a 16-byte hex id", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingRefreshTokenIdBeforeEach()

			tokenId, err := suite.sut.NewTokenId()

			assert.NoError(t, err)
			decoded, decodeErr := hex.DecodeString(tokenId)
			assert.NoError(t, decodeErr)
			assert.Len(t, decoded, 16)
		})

		t.Run("Should return a different id on every call", func(t *testing.T) {
			t.Parallel()

			suite := WhenGeneratingRefreshTokenIdBeforeEach()

			first, _ := suite.sut.NewTokenId()
			second, _ := suite.sut.NewTokenId()

			assert.NotEqual(t, first, second)
		})
	})
}
