package runtime_db

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	runtime_common_repositories "github.com/6022-labs/agentic-agent-node/src/runtime_common/repositories"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/utils"
	"github.com/ncruces/go-sqlite3/gormlite"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

	if dbSettings.Driver == enums.DatabaseDriverSqlite {
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get sql.DB: %w", err)
		}
		// SQLite has a single writer; more connections yield "database is locked".
		sqlDB.SetMaxOpenConns(1)
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
	case enums.DatabaseDriverSqlite:
		return gormlite.Open(utils.ApplySqliteDsnDefaults(url)), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}

// Use registers a gorm plugin at bootstrap, where there is no request context.
func (c *DatabaseClient) Use(plugin gorm.Plugin) error {
	return c.db.Use(plugin)
}

// GetDB returns the transaction the context carries, or the pooled connection.
func (c *DatabaseClient) GetDB(ctx context.Context, options ...runtime_common_repositories.ReadOption) *gorm.DB {
	if transaction, ok := c.TransactionFrom(ctx); ok {
		if tx, open := transaction.open(); open {
			db := tx.WithContext(ctx)
			// sqlite has no row locks; its one-connection cap serializes scopes instead.
			if slices.Contains(options, runtime_common_repositories.ToLock) && c.driver != enums.DatabaseDriverSqlite {
				db = db.Clauses(clause.Locking{Strength: "UPDATE"})
			}
			return db
		}
	}
	return c.db.WithContext(ctx)
}

func (c *DatabaseClient) GetDriver() enums.DatabaseDriver {
	return c.driver
}

func (c *DatabaseClient) Close() error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
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
