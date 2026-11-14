package settings

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

const DatabaseSettingsKey = "runtime_database"

type DatabaseSettings struct {
	Driver enums.DatabaseDriver
	Url    string
	// TlsConfig verifies the server against ca_certificate; nil leaves TLS to the DSN.
	TlsConfig *tls.Config
}

type databaseSettingsRaw struct {
	Driver        enums.DatabaseDriver `koanf:"driver"`
	Url           string               `koanf:"url"`
	CaCertificate string               `koanf:"ca_certificate"`
	TlsServerName string               `koanf:"tls_server_name"`
}

func NewDatabaseSettings(logger *zap.Logger, k *koanf.Koanf) *DatabaseSettings {
	var raw databaseSettingsRaw
	if err := k.Unmarshal(DatabaseSettingsKey, &raw); err != nil {
		logger.Fatal("failed to unmarshal runtime_database settings", zap.Error(err))
	}

	raw.Driver = enums.DatabaseDriver(strings.TrimSpace(string(raw.Driver)))
	raw.Url = strings.TrimSpace(raw.Url)
	raw.CaCertificate = strings.TrimSpace(raw.CaCertificate)
	raw.TlsServerName = strings.TrimSpace(raw.TlsServerName)

	if len(raw.Driver) == 0 {
		logger.Fatal("please set runtime_database.driver (or RUNTIME_DATABASE__DRIVER env), one of " +
			enums.GetAllDatabaseDrivers().String())
	}
	if !enums.GetAllDatabaseDrivers().Contains(raw.Driver) {
		logger.Fatal("runtime_database.driver must be one of "+enums.GetAllDatabaseDrivers().String(),
			zap.String("got", raw.Driver.String()))
	}
	if len(raw.Url) == 0 {
		logger.Fatal("please set runtime_database.url (or RUNTIME_DATABASE__URL env)")
	}

	return &DatabaseSettings{
		Driver:    raw.Driver,
		Url:       raw.Url,
		TlsConfig: raw.tlsConfig(logger),
	}
}

func (r *databaseSettingsRaw) tlsConfig(logger *zap.Logger) *tls.Config {
	if len(r.CaCertificate) == 0 {
		if len(r.TlsServerName) > 0 {
			logger.Fatal("runtime_database.tls_server_name requires runtime_database.ca_certificate")
		}
		return nil
	}
	if r.Driver == enums.DatabaseDriverSqlite {
		logger.Fatal("runtime_database.ca_certificate is not supported with the sqlite driver")
	}

	rootCAs := x509.NewCertPool()
	if !rootCAs.AppendCertsFromPEM([]byte(r.CaCertificate)) {
		logger.Fatal("runtime_database.ca_certificate holds no valid PEM certificate")
	}

	serverName := r.TlsServerName
	if len(serverName) == 0 {
		host, err := r.dsnHost()
		if err != nil {
			logger.Fatal("failed to read the host from runtime_database.url", zap.Error(err))
		}
		serverName = host
	}

	return &tls.Config{RootCAs: rootCAs, ServerName: serverName, MinVersion: tls.VersionTLS12}
}

func (r *databaseSettingsRaw) dsnHost() (string, error) {
	switch r.Driver {
	case enums.DatabaseDriverPostgres:
		config, err := pgconn.ParseConfig(r.Url)
		if err != nil {
			return "", err
		}
		return config.Host, nil
	case enums.DatabaseDriverMySQL, enums.DatabaseDriverMariaDB:
		config, err := mysql.ParseDSN(r.Url)
		if err != nil {
			return "", err
		}
		host, _, err := net.SplitHostPort(config.Addr)
		if err != nil {
			return config.Addr, nil
		}
		return host, nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s", r.Driver)
	}
}
