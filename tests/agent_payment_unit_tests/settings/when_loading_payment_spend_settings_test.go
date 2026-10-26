package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_payment/settings"
	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// fatalLogger turns a settings logger.Fatal into a panic the test can assert on.
func fatalLogger() *zap.Logger {
	return zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))
}

func TestWhenLoadingPaymentSpendSettings(t *testing.T) {
	t.Run("Given no configuration", func(t *testing.T) {
		t.Run("Should default to a bounded signing window", func(t *testing.T) {
			loaded := settings.NewPaymentSpendSettings(zap.NewNop(), koanf.New("."))

			assert.Equal(t, int64(300), loaded.ValiditySeconds)
			assert.Equal(t, int64(60), loaded.ValidAfterSkew)
		})
	})

	t.Run("Given configured values", func(t *testing.T) {
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment": map[string]any{
				"spend": map[string]any{"validity_seconds": 120, "valid_after_skew_seconds": 15},
			},
		}, "."), nil)

		t.Run("Should take them", func(t *testing.T) {
			loaded := settings.NewPaymentSpendSettings(zap.NewNop(), k)

			assert.Equal(t, int64(120), loaded.ValiditySeconds)
			assert.Equal(t, int64(15), loaded.ValidAfterSkew)
		})
	})

	t.Run("Given an out-of-range window", func(t *testing.T) {
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"payment": map[string]any{
				"spend": map[string]any{"validity_seconds": 0, "valid_after_skew_seconds": -5},
			},
		}, "."), nil)

		t.Run("Should correct it rather than refuse to boot", func(t *testing.T) {
			loaded := settings.NewPaymentSpendSettings(zap.NewNop(), k)

			assert.Equal(t, int64(300), loaded.ValiditySeconds)
			assert.Equal(t, int64(60), loaded.ValidAfterSkew)
		})
	})

	t.Run("Given the removed x402_payment namespace", func(t *testing.T) {
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"x402_payment": map[string]any{"max_amount_atomic": "1000"},
		}, "."), nil)

		t.Run("Should refuse to boot rather than ignore a stale spend limit", func(t *testing.T) {
			assert.Panics(t, func() { settings.NewPaymentSpendSettings(fatalLogger(), k) })
		})
	})

	t.Run("Given an env override", func(t *testing.T) {
		t.Setenv("PAYMENT__SPEND__VALIDITY_SECONDS", "45")

		k, err := host_configurations.LoadKoanfConfig()
		assert.NoError(t, err)

		t.Run("Should win over the defaults", func(t *testing.T) {
			loaded := settings.NewPaymentSpendSettings(zap.NewNop(), k)

			assert.Equal(t, int64(45), loaded.ValiditySeconds)
		})
	})
}
