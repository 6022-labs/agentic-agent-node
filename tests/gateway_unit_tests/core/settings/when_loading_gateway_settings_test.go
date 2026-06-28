package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/gateway/core/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingGatewaySettings(t *testing.T) {
	t.Parallel()

	t.Run("Given public_prefix is not provided", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should default to /api", func(t *testing.T) {
			t.Parallel()

			result := settings.NewGatewaySettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, "/api", result.PublicPrefix)
		})
	})

	t.Run("Given public_prefix without a leading slash", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"gateway.public_prefix": "gw",
		}, "."), nil)

		t.Run("Should normalize it to a single leading slash", func(t *testing.T) {
			t.Parallel()

			result := settings.NewGatewaySettings(zap.NewNop(), k)
			assert.Equal(t, "/gw", result.PublicPrefix)
		})
	})

	t.Run("Given public_prefix with whitespace and a trailing slash", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"gateway.public_prefix": "  /gw/  ",
		}, "."), nil)

		t.Run("Should trim whitespace and the trailing slash", func(t *testing.T) {
			t.Parallel()

			result := settings.NewGatewaySettings(zap.NewNop(), k)
			assert.Equal(t, "/gw", result.PublicPrefix)
		})
	})
}
