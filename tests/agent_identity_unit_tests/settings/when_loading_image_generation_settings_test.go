package settings_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_identity/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingImageGenerationSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given an existing image mark path", func(t *testing.T) {
		t.Parallel()

		markPath := filepath.Join(t.TempDir(), "mark.png")
		assert.NoError(t, os.WriteFile(markPath, []byte("png"), 0o600))

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"image_generation.image_mark_path": markPath,
		}, "."), nil)

		t.Run("Should load it", func(t *testing.T) {
			t.Parallel()

			result := settings.NewImageGenerationSettings(zap.NewNop(), k)

			if assert.NotNil(t, result.ImageMarkPath) {
				assert.Equal(t, markPath, *result.ImageMarkPath)
			}
		})
	})

	t.Run("Given a non-existent image mark path", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"image_generation.image_mark_path": "/does/not/exist.png",
		}, "."), nil)

		t.Run("Should disable the mark", func(t *testing.T) {
			t.Parallel()

			result := settings.NewImageGenerationSettings(zap.NewNop(), k)

			assert.Nil(t, result.ImageMarkPath)
		})
	})

	t.Run("Given a blank image mark path", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"image_generation.image_mark_path": "   ",
		}, "."), nil)

		t.Run("Should disable the mark", func(t *testing.T) {
			t.Parallel()

			result := settings.NewImageGenerationSettings(zap.NewNop(), k)

			assert.Nil(t, result.ImageMarkPath)
		})
	})

	t.Run("Given no configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")

		t.Run("Should disable the mark", func(t *testing.T) {
			t.Parallel()

			result := settings.NewImageGenerationSettings(zap.NewNop(), k)

			assert.Nil(t, result.ImageMarkPath)
		})
	})
}
