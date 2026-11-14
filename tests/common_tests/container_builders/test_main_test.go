package container_builders

import (
	"os"
	"testing"
)

// TestMain stops every shared DB container once the test process exits.
func TestMain(m *testing.M) {
	code := m.Run()
	GetPostgresContainerBuilder().stop()
	GetMySQLContainerBuilder().stop()
	GetMariaDBContainerBuilder().stop()
	GetVaultContainerBuilder().stop()
	GetAnvilContainerBuilder().stop()
	os.Exit(code)
}
