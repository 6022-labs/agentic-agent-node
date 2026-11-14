package database_client_test

import (
	"context"
	"strings"
	"testing"

	"github.com/6022-labs/agentic-agent-node/src/agent_knowledge/models/enums"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db"
	"github.com/6022-labs/agentic-agent-node/src/runtime_db/settings"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/certificate_builders"
	"github.com/6022-labs/agentic-agent-node/tests/common_tests/container_builders"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WhenCreatingDatabaseClientTestCase struct {
	name               string
	driver             enums.DatabaseDriver
	containerBuilderFn func() container_builders.TlsDatabaseContainerBuilderInterface
	// preferTlsFn rewrites a plaintext url to a TLS mode that would fall back to plaintext.
	preferTlsFn func(url string) string
	// directTlsFn rewrites a TLS url to negotiate TLS directly, which requires ALPN.
	directTlsFn          func(url string) string
	noTlsErrorMessage    string
	isSessionEncryptedFn func(db *gorm.DB) (bool, error)
}

var WhenCreatingDatabaseClientTestCases = []*WhenCreatingDatabaseClientTestCase{
	{
		name:   "PostgreSQL",
		driver: enums.DatabaseDriverPostgres,
		containerBuilderFn: func() container_builders.TlsDatabaseContainerBuilderInterface {
			return container_builders.GetPostgresContainerBuilder()
		},
		preferTlsFn: func(url string) string {
			return strings.Replace(url, "sslmode=disable", "sslmode=prefer", 1)
		},
		directTlsFn: func(url string) string {
			return url + " sslnegotiation=direct"
		},
		noTlsErrorMessage: "server refused TLS connection",
		isSessionEncryptedFn: func(db *gorm.DB) (bool, error) {
			var ssl bool
			err := db.Raw("SELECT ssl FROM pg_stat_ssl WHERE pid = pg_backend_pid()").Row().Scan(&ssl)
			return ssl, err
		},
	},
	{
		name:   "MySQL",
		driver: enums.DatabaseDriverMySQL,
		containerBuilderFn: func() container_builders.TlsDatabaseContainerBuilderInterface {
			return container_builders.GetMySQLContainerBuilder()
		},
		preferTlsFn: func(url string) string {
			return url + "&tls=preferred"
		},
		noTlsErrorMessage: "TLS requested but server does not support TLS",
		isSessionEncryptedFn: func(db *gorm.DB) (bool, error) {
			var name, cipher string
			err := db.Raw("SHOW STATUS LIKE 'Ssl_cipher'").Row().Scan(&name, &cipher)
			return cipher != "", err
		},
	},
	{
		name:   "MariaDB",
		driver: enums.DatabaseDriverMariaDB,
		containerBuilderFn: func() container_builders.TlsDatabaseContainerBuilderInterface {
			return container_builders.GetMariaDBContainerBuilder()
		},
		preferTlsFn: func(url string) string {
			return url + "&tls=preferred"
		},
		noTlsErrorMessage: "TLS requested but server does not support TLS",
		isSessionEncryptedFn: func(db *gorm.DB) (bool, error) {
			var name, cipher string
			err := db.Raw("SHOW STATUS LIKE 'Ssl_cipher'").Row().Scan(&name, &cipher)
			return cipher != "", err
		},
	},
}

type WhenCreatingDatabaseClientTestingSuite struct {
	dbName                   string
	connectionString         string
	caCertificatePem         []byte
	databaseContainerBuilder container_builders.DatabaseContainerBuilderInterface
	migrationsSettings       *settings.MigrationsSettings
}

// WhenCreatingDatabaseClientBeforeEach creates the scratch DB in the TLS container, or the plaintext one without isTls.
func WhenCreatingDatabaseClientBeforeEach(
	databaseContainerBuilder container_builders.TlsDatabaseContainerBuilderInterface,
	isTls bool,
) *WhenCreatingDatabaseClientTestingSuite {
	migrationFolderPath, err := common_tests.FindRuntimeMigrationFolderPath()
	if err != nil {
		panic(err)
	}

	suite := &WhenCreatingDatabaseClientTestingSuite{
		databaseContainerBuilder: databaseContainerBuilder,
		migrationsSettings:       &settings.MigrationsSettings{FolderPath: *migrationFolderPath},
	}
	if isTls {
		suite.dbName, suite.connectionString, suite.caCertificatePem = databaseContainerBuilder.CreateTls()
	} else {
		suite.dbName, suite.connectionString = databaseContainerBuilder.Create()
	}
	return suite
}

func WhenCreatingDatabaseClientAfterEach(suite *WhenCreatingDatabaseClientTestingSuite) {
	suite.databaseContainerBuilder.DropDatabase(suite.dbName)
}

func (s *WhenCreatingDatabaseClientTestingSuite) databaseSettings(
	driver enums.DatabaseDriver,
	caCertificatePem []byte,
	tlsServerName string,
) *settings.DatabaseSettings {
	k := koanf.New(".")
	_ = k.Load(confmap.Provider(map[string]any{
		"runtime_database.driver":          driver.String(),
		"runtime_database.url":             s.connectionString,
		"runtime_database.ca_certificate":  string(caCertificatePem),
		"runtime_database.tls_server_name": tlsServerName,
	}, "."), nil)
	return settings.NewDatabaseSettings(zap.NewNop(), k)
}

func TestWhenCreatingDatabaseClient(t *testing.T) {
	t.Parallel()

	for _, testCase := range WhenCreatingDatabaseClientTestCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			t.Run("Given a server without TLS", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse the connection", func(t *testing.T) {
					t.Parallel()

					suite := WhenCreatingDatabaseClientBeforeEach(testCase.containerBuilderFn(), false)
					defer WhenCreatingDatabaseClientAfterEach(suite)
					caCertificatePem := certificate_builders.NewCertificateAuthority().CertificatePem

					suite.connectionString = testCase.preferTlsFn(suite.connectionString)

					_, err := runtime_db.NewDatabaseClient(
						suite.databaseSettings(testCase.driver, caCertificatePem, container_builders.TlsServerName),
						suite.migrationsSettings,
						zap.NewNop(),
					)

					assert.ErrorContains(t, err, testCase.noTlsErrorMessage)
				})
			})

			if testCase.directTlsFn != nil {
				t.Run("Given direct TLS negotiation", func(t *testing.T) {
					t.Parallel()

					t.Run("Should encrypt the session", func(t *testing.T) {
						t.Parallel()

						suite := WhenCreatingDatabaseClientBeforeEach(testCase.containerBuilderFn(), true)
						defer WhenCreatingDatabaseClientAfterEach(suite)

						suite.connectionString = testCase.directTlsFn(suite.connectionString)

						client, err := runtime_db.NewDatabaseClient(
							suite.databaseSettings(
								testCase.driver,
								suite.caCertificatePem,
								container_builders.TlsServerName,
							),
							suite.migrationsSettings,
							zap.NewNop(),
						)
						if !assert.NoError(t, err) {
							return
						}
						defer client.Close()

						encrypted, err := testCase.isSessionEncryptedFn(client.GetDB(context.Background()))

						assert.NoError(t, err)
						assert.True(t, encrypted)
					})
				})
			}

			t.Run("Given a server certificate signed by another CA", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse the connection", func(t *testing.T) {
					t.Parallel()

					suite := WhenCreatingDatabaseClientBeforeEach(testCase.containerBuilderFn(), true)
					defer WhenCreatingDatabaseClientAfterEach(suite)
					caCertificatePem := certificate_builders.NewCertificateAuthority().CertificatePem

					_, err := runtime_db.NewDatabaseClient(
						suite.databaseSettings(testCase.driver, caCertificatePem, container_builders.TlsServerName),
						suite.migrationsSettings,
						zap.NewNop(),
					)

					assert.ErrorContains(t, err, "x509: certificate signed by unknown authority")
				})
			})

			// The certificate covers only TlsServerName, never the container host the url names.
			t.Run("Given no TLS server name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should refuse the connection", func(t *testing.T) {
					t.Parallel()

					suite := WhenCreatingDatabaseClientBeforeEach(testCase.containerBuilderFn(), true)
					defer WhenCreatingDatabaseClientAfterEach(suite)

					_, err := runtime_db.NewDatabaseClient(
						suite.databaseSettings(testCase.driver, suite.caCertificatePem, ""),
						suite.migrationsSettings,
						zap.NewNop(),
					)

					assert.ErrorContains(t, err, "x509: certificate is valid for "+container_builders.TlsServerName)
				})
			})

			t.Run("Given the server CA and TLS server name", func(t *testing.T) {
				t.Parallel()

				t.Run("Should apply all migrations without error", func(t *testing.T) {
					t.Parallel()

					suite := WhenCreatingDatabaseClientBeforeEach(testCase.containerBuilderFn(), true)
					defer WhenCreatingDatabaseClientAfterEach(suite)

					client, err := runtime_db.NewDatabaseClient(
						suite.databaseSettings(
							testCase.driver,
							suite.caCertificatePem,
							container_builders.TlsServerName,
						),
						suite.migrationsSettings,
						zap.NewNop(),
					)
					if !assert.NoError(t, err) {
						return
					}
					defer client.Close()

					assert.NoError(t, client.Migrate())
				})

				t.Run("Should encrypt the session", func(t *testing.T) {
					t.Parallel()

					suite := WhenCreatingDatabaseClientBeforeEach(testCase.containerBuilderFn(), true)
					defer WhenCreatingDatabaseClientAfterEach(suite)

					client, err := runtime_db.NewDatabaseClient(
						suite.databaseSettings(
							testCase.driver,
							suite.caCertificatePem,
							container_builders.TlsServerName,
						),
						suite.migrationsSettings,
						zap.NewNop(),
					)
					if !assert.NoError(t, err) {
						return
					}
					defer client.Close()

					encrypted, err := testCase.isSessionEncryptedFn(client.GetDB(context.Background()))

					assert.NoError(t, err)
					assert.True(t, encrypted)
				})
			})
		})
	}
}
