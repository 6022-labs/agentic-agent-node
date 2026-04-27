package settings

import (
	"os"
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const MigrationsSettingsKey = "agent_knowledge_migrations"

type MigrationsSettings struct {
	FolderPath string `koanf:"folder_path"`
}

func NewMigrationsSettings(logger *zap.Logger, k *koanf.Koanf) *MigrationsSettings {
	var settings MigrationsSettings
	if err := k.Unmarshal(MigrationsSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal agent_knowledge_migrations settings", zap.Error(err))
	}

	settings.FolderPath = strings.TrimSpace(settings.FolderPath)
	if len(settings.FolderPath) == 0 {
		settings.FolderPath = "./src/agent_knowledge_db/migrations"
		logger.Warn("agent_knowledge_migrations.folder_path not set, using default", zap.String("path", settings.FolderPath))
	}

	if _, err := os.Stat(settings.FolderPath); os.IsNotExist(err) {
		logger.Fatal("agent_knowledge_migrations folder does not exist", zap.String("path", settings.FolderPath), zap.Error(err))
	}

	return &settings
}
