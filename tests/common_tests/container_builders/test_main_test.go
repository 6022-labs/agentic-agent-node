package container_builders

import (
	"os"
	"testing"
)

// TestMain stops every shared DB container once the test process exits.
func TestMain(m *testing.M) {
	code := m.Run()
	GetPostgresContainerBuilder().stop()
	GetPostgresTlsContainerBuilder().stop()
	GetMySQLContainerBuilder().stop()
	GetMySQLTlsContainerBuilder().stop()
	GetMariaDBContainerBuilder().stop()
	GetMariaDBTlsContainerBuilder().stop()
	GetVaultContainerBuilder().stop()
	GetAnvilContainerBuilder().stop()
	os.Exit(code)
}
