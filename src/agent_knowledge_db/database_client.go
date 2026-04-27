package agent_knowledge_db

import (
	"fmt"
	"sync"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/settings"
	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge_db/utils"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

// goose forbids concurrent migrations on the same DB.
var gooseMigrationMu sync.Mutex

type DatabaseClient struct {
	db                 *gorm.DB
	driver             enums.DatabaseDriver
	migrationsSettings *settings.MigrationsSettings
}

func NewDatabaseClient(
	dbSettings *settings.DatabaseSettings,
	migrationsSettings *settings.MigrationsSettings,
	logger *zap.Logger,
) (*DatabaseClient, error) {
	dialector, err := dialectorFor(dbSettings.Driver, dbSettings.Url)
	if err != nil {
		return nil, err
	}

	gormLogger := zapgorm2.New(logger)
	gormLogger.IgnoreRecordNotFoundError = true

	db, err := gorm.Open(dialector, &gorm.Config{Logger: gormLogger})
	if err != nil {
		return nil, fmt.Errorf("failed to open agent_knowledge database: %w", err)
	}

	return &DatabaseClient{
		db:                 db,
		driver:             dbSettings.Driver,
		migrationsSettings: migrationsSettings,
	}, nil
}

func dialectorFor(driver enums.DatabaseDriver, url string) (gorm.Dialector, error) {
	switch driver {
	case enums.DatabaseDriverPostgres:
		return postgres.Open(url), nil
	case enums.DatabaseDriverMySQL, enums.DatabaseDriverMariaDB:
		return mysql.Open(url), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}

func (c *DatabaseClient) GetDB() *gorm.DB {
	return c.db
}

func (c *DatabaseClient) GetDriver() enums.DatabaseDriver {
	return c.driver
}

func (c *DatabaseClient) HealthCheck() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (c *DatabaseClient) Migrate() error {
	gooseMigrationMu.Lock()
	defer gooseMigrationMu.Unlock()

	sqlDB, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	if err := goose.SetDialect(c.driver.ToGooseDialect()); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	folder := utils.GetMigrationFolderPath(c.migrationsSettings.FolderPath, c.driver.String())
	if err := goose.Up(sqlDB, folder); err != nil {
		return fmt.Errorf("goose migration failed: %w", err)
	}
	return nil
}
