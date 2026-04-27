package settings_test

import (
	"testing"

	knowledge_settings "github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingMigrationsSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given an existing folder_path", func(t *testing.T) {
		t.Parallel()
		// t.TempDir is created + cleaned up automatically; satisfies the os.Stat check.
		dir := t.TempDir()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_knowledge_migrations.folder_path": dir,
		}, "."), nil)

		t.Run("Should load the path verbatim", func(t *testing.T) {
			t.Parallel()
			result := knowledge_settings.NewMigrationsSettings(zap.NewNop(), k)
			assert.Equal(t, dir, result.FolderPath)
		})
	})

	t.Run("Given folder_path has surrounding whitespace", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_knowledge_migrations.folder_path": "  " + dir + "  ",
		}, "."), nil)

		t.Run("Should trim it", func(t *testing.T) {
			t.Parallel()
			result := knowledge_settings.NewMigrationsSettings(zap.NewNop(), k)
			assert.Equal(t, dir, result.FolderPath)
		})
	})
}
