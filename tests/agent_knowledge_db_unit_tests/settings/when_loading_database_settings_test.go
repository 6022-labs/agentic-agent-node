package settings_test

import (
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	knowledge_settings "github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/settings"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWhenLoadingDatabaseSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a postgres configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_knowledge_database.driver": "postgres",
			"agent_knowledge_database.url":    "postgres://u:p@localhost:5432/db?sslmode=disable",
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, enums.DatabaseDriverPostgres, result.Driver)
			assert.Equal(t, "postgres://u:p@localhost:5432/db?sslmode=disable", result.Url)
		})
	})

	t.Run("Given a mysql configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_knowledge_database.driver": "mysql",
			"agent_knowledge_database.url":    "user:pass@tcp(localhost:3306)/db?parseTime=true",
		}, "."), nil)

		t.Run("Should accept the mysql driver", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, enums.DatabaseDriverMySQL, result.Driver)
		})
	})

	t.Run("Given a mariadb configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_knowledge_database.driver": "mariadb",
			"agent_knowledge_database.url":    "user:pass@tcp(localhost:3306)/db?parseTime=true",
		}, "."), nil)

		t.Run("Should accept the mariadb driver", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, enums.DatabaseDriverMariaDB, result.Driver)
		})
	})

	t.Run("Given values have surrounding whitespace", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"agent_knowledge_database.driver": "  postgres  ",
			"agent_knowledge_database.url":    "  postgres://x  ",
		}, "."), nil)

		t.Run("Should trim driver and url", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, enums.DatabaseDriverPostgres, result.Driver)
			assert.Equal(t, "postgres://x", result.Url)
		})
	})
}
