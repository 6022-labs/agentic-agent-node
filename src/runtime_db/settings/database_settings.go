package settings

import (
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const DatabaseSettingsKey = "runtime_database"

type DatabaseSettings struct {
	Driver enums.DatabaseDriver `koanf:"driver"`
	Url    string               `koanf:"url"`
}

func NewDatabaseSettings(logger *zap.Logger, k *koanf.Koanf) *DatabaseSettings {
	var settings DatabaseSettings
	if err := k.Unmarshal(DatabaseSettingsKey, &settings); err != nil {
		logger.Fatal("failed to unmarshal runtime_database settings", zap.Error(err))
	}

	settings.Driver = enums.DatabaseDriver(strings.TrimSpace(string(settings.Driver)))
	settings.Url = strings.TrimSpace(settings.Url)

	if len(settings.Driver) == 0 {
		logger.Fatal("please set runtime_database.driver (or RUNTIME_DATABASE__DRIVER env), one of " +
			enums.GetAllDatabaseDrivers().String())
	}
	if !enums.GetAllDatabaseDrivers().Contains(settings.Driver) {
		logger.Fatal("runtime_database.driver must be one of "+enums.GetAllDatabaseDrivers().String(),
			zap.String("got", settings.Driver.String()))
	}
	if len(settings.Url) == 0 {
		logger.Fatal("please set runtime_database.url (or RUNTIME_DATABASE__URL env)")
	}

	return &settings
}
