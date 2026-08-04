package settings

import (
	"strings"

	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const OsSettingsKey = "os"

const defaultWorkdir = "/workspace"

// OsSettings covers the machine the agent acts on, so anything reaching it shares this — commands today, files later.
// CommandUser is empty by default: dropping to an account that does not exist would be worse than not dropping.
type OsSettings struct {
	Workdir     string `koanf:"workdir"`
	CommandUser string `koanf:"command_user"`
}

func NewOsSettings(logger *zap.Logger, k *koanf.Koanf) *OsSettings {
	settings := OsSettings{
		Workdir: defaultWorkdir,
	}
	if err := k.Unmarshal(OsSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal os settings", zap.Error(err))
	}

	settings.CommandUser = strings.TrimSpace(settings.CommandUser)

	settings.Workdir = strings.TrimSpace(settings.Workdir)
	if settings.Workdir == "" {
		logger.Warn("invalid os.workdir, defaulting to /workspace")
		settings.Workdir = defaultWorkdir
	}

	return &settings
}
