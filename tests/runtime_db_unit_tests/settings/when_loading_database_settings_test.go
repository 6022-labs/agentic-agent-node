package settings_test

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/common/host_configurations"
	knowledge_settings "github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/certificate_builders"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestWhenLoadingDatabaseSettings(t *testing.T) {
	t.Parallel()

	t.Run("Given a postgres configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver": "postgres",
			"runtime_database.url":    "postgres://u:p@localhost:5432/db?sslmode=disable",
		}, "."), nil)

		t.Run("Should load every field as provided", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)

			assert.NotNil(t, result)
			assert.Equal(t, enums.DatabaseDriverPostgres, result.Driver)
			assert.Equal(t, "postgres://u:p@localhost:5432/db?sslmode=disable", result.Url)
		})
	})

	t.Run("Given a mysql configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver": "mysql",
			"runtime_database.url":    "user:pass@tcp(localhost:3306)/db?parseTime=true",
		}, "."), nil)

		t.Run("Should accept the mysql driver", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, enums.DatabaseDriverMySQL, result.Driver)
		})
	})

	t.Run("Given a mariadb configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver": "mariadb",
			"runtime_database.url":    "user:pass@tcp(localhost:3306)/db?parseTime=true",
		}, "."), nil)

		t.Run("Should accept the mariadb driver", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, enums.DatabaseDriverMariaDB, result.Driver)
		})
	})

	t.Run("Given a sqlite configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver": "sqlite",
			"runtime_database.url":    "/data/agent.db",
		}, "."), nil)

		t.Run("Should accept the sqlite driver", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, enums.DatabaseDriverSqlite, result.Driver)
		})
	})

	t.Run("Given padded values", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver": "  postgres  ",
			"runtime_database.url":    "  postgres://x  ",
		}, "."), nil)

		t.Run("Should trim driver and url", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, enums.DatabaseDriverPostgres, result.Driver)
			assert.Equal(t, "postgres://x", result.Url)
		})
	})

	t.Run("Given no CA certificate", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver": "postgres",
			"runtime_database.url":    "postgres://u:p@localhost:5432/db?sslmode=require",
		}, "."), nil)

		t.Run("Should leave TLS to the url", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Nil(t, result.TlsConfig)
		})
	})

	t.Run("Given a TLS server name without a CA certificate", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver":          "postgres",
			"runtime_database.url":             "postgres://u:p@localhost:5432/db",
			"runtime_database.tls_server_name": "stack-postgres.internal",
		}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() {
				knowledge_settings.NewDatabaseSettings(logger, k)
			})
		})
	})

	t.Run("Given an invalid CA certificate", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver":         "postgres",
			"runtime_database.url":            "postgres://u:p@localhost:5432/db",
			"runtime_database.ca_certificate": "not a certificate",
		}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() {
				knowledge_settings.NewDatabaseSettings(logger, k)
			})
		})
	})

	t.Run("Given a CA certificate on a sqlite configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver":         "sqlite",
			"runtime_database.url":            "/data/agent.db",
			"runtime_database.ca_certificate": string(certificate_builders.NewCertificateAuthority().CertificatePem),
		}, "."), nil)

		t.Run("Should fail fast", func(t *testing.T) {
			t.Parallel()

			logger := zap.NewNop().WithOptions(zap.WithFatalHook(zapcore.WriteThenPanic))

			assert.Panics(t, func() {
				knowledge_settings.NewDatabaseSettings(logger, k)
			})
		})
	})

	t.Run("Given a CA certificate on a postgres configuration", func(t *testing.T) {
		t.Parallel()

		caCertificatePem := certificate_builders.NewCertificateAuthority().CertificatePem
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver":         "postgres",
			"runtime_database.url":            "postgres://u:p@db.example.com:5432/db?sslmode=disable",
			"runtime_database.ca_certificate": string(caCertificatePem),
		}, "."), nil)

		t.Run("Should trust only the CA", func(t *testing.T) {
			t.Parallel()

			expected := x509.NewCertPool()
			expected.AppendCertsFromPEM(caCertificatePem)

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.True(t, expected.Equal(result.TlsConfig.RootCAs))
		})

		t.Run("Should require TLS 1.2", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, uint16(tls.VersionTLS12), result.TlsConfig.MinVersion)
		})

		t.Run("Should default the server name to the url host", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, "db.example.com", result.TlsConfig.ServerName)
		})
	})

	t.Run("Given a CA certificate on a mysql configuration", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver":         "mysql",
			"runtime_database.url":            "user:pass@tcp(db.example.com:3306)/db?parseTime=true",
			"runtime_database.ca_certificate": string(certificate_builders.NewCertificateAuthority().CertificatePem),
		}, "."), nil)

		t.Run("Should default the server name to the url host", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, "db.example.com", result.TlsConfig.ServerName)
		})
	})

	t.Run("Given a CA certificate and a TLS server name", func(t *testing.T) {
		t.Parallel()

		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{
			"runtime_database.driver":          "mariadb",
			"runtime_database.url":             "user:pass@tcp(proxy.example.com:3306)/db?parseTime=true",
			"runtime_database.ca_certificate":  string(certificate_builders.NewCertificateAuthority().CertificatePem),
			"runtime_database.tls_server_name": "  stack-mariadb.internal  ",
		}, "."), nil)

		t.Run("Should use the TLS server name", func(t *testing.T) {
			t.Parallel()

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)
			assert.Equal(t, "stack-mariadb.internal", result.TlsConfig.ServerName)
		})
	})
}

func TestWhenLoadingDatabaseSettingsFromEnv(t *testing.T) {
	t.Run("Given an env override", func(t *testing.T) {
		caCertificatePem := certificate_builders.NewCertificateAuthority().CertificatePem
		t.Setenv("RUNTIME_DATABASE__DRIVER", "postgres")
		t.Setenv("RUNTIME_DATABASE__URL", "postgres://u:p@proxy.example.com:5432/db")
		t.Setenv("RUNTIME_DATABASE__CA_CERTIFICATE", string(caCertificatePem))
		t.Setenv("RUNTIME_DATABASE__TLS_SERVER_NAME", "stack-postgres.internal")

		k, err := host_configurations.LoadKoanfConfig()
		assert.NoError(t, err)

		t.Run("Should win over the defaults", func(t *testing.T) {
			expected := x509.NewCertPool()
			expected.AppendCertsFromPEM(caCertificatePem)

			result := knowledge_settings.NewDatabaseSettings(zap.NewNop(), k)

			assert.True(t, expected.Equal(result.TlsConfig.RootCAs))
			assert.Equal(t, "stack-postgres.internal", result.TlsConfig.ServerName)
		})
	})
}
