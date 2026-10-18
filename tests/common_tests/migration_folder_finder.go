package common_tests

import (
	"os"
	"path/filepath"
)

// findModuleRoot walks up from cwd to the go.mod directory so tests resolve assets from any package directory.
func findModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func findMigrationFolderPath(subPath string) (*string, error) {
	root, err := findModuleRoot()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(root, subPath)
	return &p, nil
}

// FindRuntimeMigrationFolderPath returns the absolute path to the runtime_db migrations folder.
func FindRuntimeMigrationFolderPath() (*string, error) {
	return findMigrationFolderPath("src/runtime_db/migrations")
}
