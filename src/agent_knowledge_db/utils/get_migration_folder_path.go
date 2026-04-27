package utils

import "path/filepath"

// GetMigrationFolderPath returns the per-driver migrations subfolder under
// the base path. Each driver has its own SQL file set because Postgres and
// MySQL diverge on type names + index syntax.
func GetMigrationFolderPath(basePath string, driver string) string {
	return filepath.Join(basePath, driver)
}
