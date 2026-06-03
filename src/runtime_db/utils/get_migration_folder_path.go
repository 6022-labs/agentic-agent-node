package utils

import "path/filepath"

// GetMigrationFolderPath returns the per-driver migrations subfolder; Postgres
// and MySQL need separate SQL (diverging type names + index syntax).
func GetMigrationFolderPath(basePath string, driver string) string {
	return filepath.Join(basePath, driver)
}
