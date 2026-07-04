package container_builders

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

var sqliteBuilder *SqliteContainerBuilder

// SqliteContainerBuilder needs no container: each scratch DB is a temp file.
type SqliteContainerBuilder struct {
	dirMu   sync.Mutex
	dir     string
	filesMu sync.Mutex
	files   map[string]string
}

func GetSqliteContainerBuilder() *SqliteContainerBuilder {
	if sqliteBuilder == nil {
		sqliteBuilder = &SqliteContainerBuilder{files: map[string]string{}}
	}
	return sqliteBuilder
}

func (b *SqliteContainerBuilder) Create() (string, string) {
	b.dirMu.Lock()
	if b.dir == "" {
		dir, err := os.MkdirTemp("", "sqlite_tests_")
		if err != nil {
			log.Fatalf("failed to create sqlite temp dir: %v", err)
		}
		b.dir = dir
	}
	b.dirMu.Unlock()

	dbName := "test_" + uuid.New().String()[:8]
	path := filepath.Join(b.dir, dbName+".db")

	b.filesMu.Lock()
	b.files[dbName] = path
	b.filesMu.Unlock()

	return dbName, path
}

func (b *SqliteContainerBuilder) DropDatabase(dbName string) {
	b.filesMu.Lock()
	defer b.filesMu.Unlock()
	if path, ok := b.files[dbName]; ok {
		for _, suffix := range []string{"", "-wal", "-shm"} {
			_ = os.Remove(path + suffix)
		}
		delete(b.files, dbName)
	}
}
